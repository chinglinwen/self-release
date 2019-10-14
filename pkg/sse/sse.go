package sse

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/chinglinwen/log"

	"github.com/mohae/deepcopy"
)

// var brokers = []Broker{}

var brokerMaps sync.Map

// A single Broker will be created in this program. It is responsible
// for keeping a list of which clients (browsers) are currently attached
// and broadcasting events (messages) to those clients.
//
type Broker struct {
	Project string // project
	Key     string // unique key
	Branch  string
	Event   *EventInfo

	Retry int

	// Channel into which messages are pushed to be broadcast out
	// to attahed clients.
	//
	Messages chan string `json:"-"`

	PReader *io.PipeReader `json:"-"`
	PWriter *io.PipeWriter `json:"-"`

	// Create a map of clients, the keys of the map are the channels
	// over which we can push messages to attached clients.  (The values
	// are just booleans and are meaningless.)
	//
	clients map[chan string]bool `json:"-"`

	// Channel into which new clients can be pushed
	//
	newClients chan chan string `json:"-"`

	// Channel into which disconnected clients should be pushed
	//
	defunctClients chan chan string `json:"-"`

	ExistMsg   []string
	CreateTime string
	Stored     bool
}

// this type now relate to k8s project object
type EventInfo struct {
	Project string `json:"project,omitempty"` // event.Project.PathWithNamespace
	Branch  string `json:"version,omitempty"` // parseBranch(event.Ref)

	Env string `json:"env,omitempty"` // auto detect

	UserName  string `json:"userName,omitempty"`
	UserEmail string `json:"userEmail,omitempty"`
	Message   string `json:"releaseMessage,omitempty"`
	Time      string `json:"releaseAt,omitempty"`

	CommitID string `json:"-"`
}

func ParseEventInfoJson(body string) (event *EventInfo, err error) {
	event = &EventInfo{}
	err = json.Unmarshal([]byte(body), event)
	if err != nil {
		return
	}
	return
}

// GetInfo to satisfy eventer
func (e *EventInfo) GetInfo() (event *EventInfo, err error) {
	return e, nil
}

const TimeLayout = "2006-1-2_15:04:05"

type option struct {
	key      string
	event    EventInfo
	rollback string
}

func SetKey(key string) func(*option) {
	return func(o *option) {
		o.key = key
	}
}
func SetRollback(rollback string) func(*option) {
	return func(o *option) {
		o.rollback = rollback
	}
}
func SetEventInfo(event EventInfo) func(*option) {
	return func(o *option) {
		o.event = event
	}
}

// how to log everytime's log? history logs?
// store history somewhere(in fs), read it later?
func New(project, branch string, options ...func(*option)) (b *Broker) {
	c := &option{}
	for _, op := range options {
		op(c)
	}
	if c.key == "" {
		c.key = strings.Replace(fmt.Sprintf("%v-%v", project, branch), "/", "-", -1)
	}
	log.Printf("created new logs key: %v", c.key)

	pr, pw := io.Pipe()

	x, ok := brokerMaps.Load(c.key)
	if ok {
		b, ok = x.(*Broker)
		if !ok {
			log.Println("convert from brockerMaps error for ", c.key)
			return
		}
	} else {
		b = &Broker{
			Key:            c.key,
			Project:        project,
			Branch:         branch,
			Messages:       make(chan string, 100),
			PReader:        pr,
			PWriter:        pw,
			clients:        make(map[chan string]bool),
			newClients:     make(chan (chan string)),
			defunctClients: make(chan (chan string)),
			CreateTime:     time.Now().Format(TimeLayout),
		}
	}

	b.Start()
	// Generate a constant stream of events that get pushed
	// into the Broker's messages channel and are then broadcast
	// out to any clients that are attached.

	brokerMaps.Store(c.key, b)

	return b
}

func NewExist(b *Broker) (bnew *Broker) {
	// c := &option{}
	// for _, op := range options {
	// 	op(c)
	// }
	// if c.key == "" {
	// 	c.key = strings.Replace(fmt.Sprintf("%v-%v", project, branch), "/", "-", -1)
	// }

	// b.Key=b.Key+".retry"
	pr, pw := io.Pipe()

	key := b.Key
	project := b.Project
	branch := b.Branch

	x, ok := brokerMaps.Load(key)
	if ok {
		bnew, ok = x.(*Broker)
		if !ok {
			log.Println("convert from brockerMaps error for ", key)
			return
		}
	} else {
		bnew = &Broker{
			// Key:            fmt.Sprintf("%v-retry%v", key, b.Retry),
			Key:            key,
			Project:        project,
			Branch:         branch,
			Messages:       make(chan string),
			PReader:        pr,
			PWriter:        pw,
			clients:        make(map[chan string]bool),
			newClients:     make(chan (chan string)),
			defunctClients: make(chan (chan string)),
			CreateTime:     time.Now().Format(TimeLayout),
			Event:          b.Event,
			ExistMsg:       []string{fmt.Sprintf("<h2>retried at %v</h2>\n", time.Now().Format(TimeLayout))},
			Retry:          b.Retry + 1,
		}
	}
	if bnew.Project == "" {
		log.Printf("got empty prjoect for broker, should not happen\nbroker: %#v\n", bnew)
	}

	bnew.Start()
	// Generate a constant stream of events that get pushed
	// into the Broker's messages channel and are then broadcast
	// out to any clients that are attached.

	brokerMaps.Store(key, bnew)

	return bnew
}

func GetBrokers() (bs []*Broker, err error) {
	bs = GetBrokersFromMem()
	dbs, err := GetBrokersFromDisk()
	if err != nil {
		return
	}
	bs = append(bs, dbs...)
	if bs == nil || len(bs) == 0 {
		err = fmt.Errorf("no anything found")
		return
	}
	return
}

func GetBrokersFromMem() []*Broker {
	// spew.Dump("brokerMaps", brokerMaps)
	bs := []*Broker{}
	brokerMaps.Range(func(k, v interface{}) bool {
		// spew.Dump("k", k, v)
		if b, ok := v.(*Broker); ok {
			bs = append(bs, b)
			// bs[k.(string)] = b
		} else {
			log.Println("cast back to broker error", v)
		}

		return true
	})
	if bs == nil {
		return nil
	}
	// spew.Dump("bs", bs)
	log.Printf("got %v brokers from mem", len(bs))

	// for _, v := range bs {
	// 	log.Printf("ev: %v", v.Event)
	// }

	// sort.Slice(bs, func(i, j int) bool {
	// 	if bs[i].Event == nil || bs[j].Event == nil {
	// 		log.Printf("got nil event broker, should not happen")
	// 		return false
	// 	}
	// 	// spew.Dump("i-ev: %v,j-ev: %v", bs[i], bs[j])
	// 	return bs[i].Event.Time.Before(bs[j].Event.Time) // recent first?
	// })
	// spew.Dump("bs", bs)
	return bs
}

func GetBrokerFromKey(key string) (b *Broker, err error) {
	bs, err := GetBrokers()
	if err != nil {
		return
	}
	for _, v := range bs {
		if v.Key == key {
			b = v
			return
		}
	}
	return
}

func GetBrokerFromPerson(name string) (b *Broker, err error) {
	bs, err := GetBrokersFromDisk() // no includes of mem brockers
	if err != nil {
		return
	}
	if bs == nil {
		err = fmt.Errorf("no any project")
		return
	}

	// for _, v := range bs {
	// 	fmt.Printf("key: %v\n", v.Key)
	// }
	// spew.Dump("bs", bs)
	for _, v := range bs {
		// fmt.Printf("key: %v\n", v.Key)
		// spew.Dump("v", v)
		if v.Event == nil {
			continue
		}
		if v.Event.UserName == name {
			b = v
			return
		}
	}
	if b == nil {
		err = fmt.Errorf("%v haven't build project yet", name)
		return
	}

	return
}

func (b *Broker) GetExistMsg() (existmsg string) {
	for _, v := range b.ExistMsg {
		existmsg = fmt.Sprintf("%v%v\n", existmsg, v)
	}
	return
}

var builderLock map[string]bool // mutex for operation

// is this need to base on project only?
func Lock(project, branch string) (err error) {
	if builderLock == nil {
		builderLock = make(map[string]bool)
	}
	k := fmt.Sprintf("%v:%v", project, branch)
	if v, ok := builderLock[k]; ok && v {
		err = fmt.Errorf("operation is in running, try later")
		return
	}
	builderLock[k] = true
	return
}

func UnLock(project, branch string) (err error) {
	k := fmt.Sprintf("%v:%v", project, branch)
	if v, ok := builderLock[k]; !ok || !v {
		err = fmt.Errorf("there's no lock")
		return
	}
	builderLock[k] = false
	return
}

func (b *Broker) Close() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("panic happened", r)
		}

	}()

	log.Println("closing brocker for ", b.Project, b.Branch)
	if b.PWriter != nil {
		b.PWriter.Close()
		log.Debug.Println("writer close ok")
	}

	// copy as backup, the name is the same? how to distinguish later
	// key := b.Project + "." + b.CreateTime

	key := strings.Replace(fmt.Sprintf("%v:%v-%v", b.Project, b.Branch, b.CreateTime), "/", "-", -1)

	b1 := deepcopy.Copy(b)
	newb, _ := b1.(*Broker)

	newb.Key = key
	newb.Stored = true

	// store to local distk too? not to store in memory, because it will lost, and occupy memory
	// brokerMaps.Store(key, newb)
	err := WriteFile(key, newb)
	if err != nil {
		log.Printf("close broker and backup as %v, err: %v\n", key, err)
		return
	}
	log.Printf("close broker and backup as %v ok\n", key)
}

// This Broker method starts a new goroutine.  It handles
// the addition & removal of clients, as well as the broadcasting
// of messages out to clients that are currently attached.
//
func (b *Broker) Start() {
	if b == nil {
		log.Println("nil brocker for")
		return
	}
	log.Println("starting brocker for ", b.Key)

	// existMsg := []string{}

	// we may need make this for every clients

	// go func() {
	// 	log.Println("try reading msg into brocker for ", b.Name)
	// 	// spew.Dump(b.PReader)
	// 	// pretty(b.PReader)
	// 	scanner := bufio.NewScanner(b.PReader)
	// 	for scanner.Scan() {
	// 		msg := scanner.Text()
	// 		log.Printf("%v --> msg: %q \n", b.Name, msg)
	// 		existMsg = append(existMsg, msg)
	// 		b.Messages <- msg
	// 	}
	// }()

	go func() {
		log.Println("try reading msg into brocker for ", b.Key)

		scanner := bufio.NewScanner(b.PReader)
		for scanner.Scan() {
			msg := scanner.Text()
			// log.Printf("%v --> msg: %q \n", b.Key, msg)
			b.ExistMsg = append(b.ExistMsg, msg)
			b.Messages <- msg
		}

		// p := make([]byte, 256) // make it long enough to not split lines
		// for {
		// 	n, err := b.PReader.Read(p)
		// 	if err == io.EOF {
		// 		break
		// 	}
		// 	msg := string(p[:n])
		// 	// log.Printf("%v --> msg: %q \n", b.Key, msg)
		// 	b.ExistMsg = append(b.ExistMsg, msg)
		// 	b.Messages <- msg
		// }

		// store msg into local fs, for later retrive? or just stay with branch
	}()

	// Start a goroutine
	//
	go func() {

		// Loop endlessly
		//
		for {

			// Block until we receive from one of the
			// three following channels.
			select {

			case s := <-b.newClients:

				// There is a new client attached and we
				// want to start sending them messages.
				b.clients[s] = true
				// log.Println("Added new client")

				// read existing msg, send it
				for _, v := range b.ExistMsg {
					s <- v
				}

			case s := <-b.defunctClients:

				// A client has dettached and we want to
				// stop sending them messages.
				delete(b.clients, s)
				close(s)

				log.Println("Removed client")

			case msg := <-b.Messages: // how to include old msg?

				// how to send earlier log, if later attached

				// There is a new message to send.  For each
				// attached client, push the new message
				// into the client's message channel.
				for s := range b.clients {
					s <- msg
				}
				// log.Printf("Broadcast message to %d clients", len(b.clients))
			}
		}
	}()
}

// This Broker method handles and HTTP request at the "/events/" URL.
//
func SSEHandler(w http.ResponseWriter, r *http.Request) {
	// spew.Dump(r.Header)

	// pretty(r)

	var err error
	// Make sure that the writer supports flushing.
	//
	f, ok := w.(http.Flusher)
	if !ok {
		err = fmt.Errorf("Streaming unsupported!")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	project := r.FormValue("project")
	branch := r.FormValue("branch")
	key := r.FormValue("key")

	if project == "" && key == "" {
		err = fmt.Errorf("project query value is empty")
		fmt.Fprintf(w, "%v\n", err)
		log.Println(err)
		return
	}
	log.Printf("got project: %v, key: %v for eventsources\n", project, key)

	if branch == "" {
		branch = "develop"
	}

	needcreate := true
	if key == "" {
		needcreate = false
	}
	// bname := fmt.Sprintf("%v-%v", project, branch)

	var b *Broker
	x, ok := brokerMaps.Load(key)
	if ok {
		b, ok = x.(*Broker)
		if !ok {
			err = fmt.Errorf("convert back broker error")
			fmt.Fprintf(w, "%v\n", err)
			log.Println(err)
			return
		}
		log.Printf("got existing broker for %v\n", key)
	} else {
		if needcreate {
			// err = fmt.Errorf("project doesn't exist yet")
			// fmt.Fprintf(w, "%v\n", err)
			// log.Println(err)
			// return
			b = New(project, branch)
			// spew.Dump("broker", b)
			log.Printf("created broker for %v\n", project)
		} else {
			log.Printf("not created broker for %v, it should exist\n", key)
		}
	}

	if b == nil {
		log.Println("got empty broker")
		return
	}

	// Create a new channel, over which the broker can
	// send this client messages.
	messageChan := make(chan string)

	// Add this client to the map of those that should
	// receive updates
	b.newClients <- messageChan

	// Listen to the closing of the http connection via the CloseNotifier
	notify := w.(http.CloseNotifier).CloseNotify()
	go func() {
		<-notify
		// Remove this client from the map of attached clients
		// when `EventHandler` exits.
		b.defunctClients <- messageChan
		log.Println("HTTP connection just closed.")
	}()

	// Set the headers related to event streaming.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")

	// Don't close the connection, instead loop endlessly.
	for {

		// Read from our messageChan.
		msg, open := <-messageChan

		if !open {
			// If our messageChan was closed, this means that the client has
			// disconnected.
			break
		}

		// Write to the ResponseWriter, `w`.
		fmt.Fprintf(w, "data: %s\n\n", msg)

		// Flush the response.  This is only possible if
		// the repsonse supports streaming.
		f.Flush()
	}

	// Done.
	// log.Println("Finished HTTP request at ", r.URL.Path)
}

func pretty(a interface{}) {
	b, _ := json.MarshalIndent(a, "", "  ")
	fmt.Println("pretty", string(b))
}

// // Handler for the main page, which we wire up to the
// // route at "/" below in `main`.
// //
// func UIHandler(w http.ResponseWriter, r *http.Request) {

// 	// Did you know Golang's ServeMux matches only the
// 	// prefix of the request URL?  It's true.  Here we
// 	// insist the path is just "/".
// 	// if r.URL.Path != "/" {
// 	// 	w.WriteHeader(http.StatusNotFound)
// 	// 	return
// 	// }

// 	// Read in the template with our SSE JavaScript code.
// 	t, err := template.ParseFiles("pkg/sse/index.html")
// 	if err != nil {
// 		log.Fatal("WTF dude, error parsing your template.")

// 	}
// 	log.Println("parsed template")

// 	// Render the template, writing to `w`.
// 	t.Execute(w, "friend")

// 	// Done.
// 	log.Println("Finished HTTP request at", r.URL.Path)
// }

// // Main routine
// //
// func main() {

// 	// Make a new Broker instance
// 	b := &Broker{
// 		make(map[chan string]bool),
// 		make(chan (chan string)),
// 		make(chan (chan string)),
// 		make(chan string),
// 	}

// 	// Start processing events
// 	b.Start()

// 	// Make b the HTTP handler for "/events/".  It can do
// 	// this because it has a ServeHTTP method.  That method
// 	// is called in a separate goroutine for each
// 	// request to "/events/".
// 	http.Handle("/events/", b)

// 	// When we get a request at "/", call `handler`
// 	// in a new goroutine.
// 	http.Handle("/", http.HandlerFunc(handler))

// 	// Start the server and listen forever on port 8000.
// 	http.ListenAndServe(":8000", nil)
// }
