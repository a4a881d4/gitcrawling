package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	//"strconv"
	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/emicklei/go-restful"
	swagger "github.com/emicklei/go-restful-swagger12"
)

type Refs struct {
	Hash, TreeHash string
}
type DBGeter interface {
	GetRawRef(h []byte, cb func([]byte) error) error
}
type RefResource struct {
	ref DBGeter
}

func NewRefResource(r DBGeter) *RefResource {
	return &RefResource{r}
}
func (u RefResource) Register(container *restful.Container) {
	ws := new(restful.WebService)

	ws.
		Path("/r").
		Doc("Hash of Ref").
		Consumes(restful.MIME_XML, restful.MIME_JSON).
		Produces(restful.MIME_JSON, restful.MIME_XML)

	ws.Route(ws.GET("/{hash}").To(u.findRefs).
		// docs
		Doc("get a refs").
		Operation("findRefs").
		Param(ws.PathParameter("hash", "hash of the refs").DataType("string")).
		Writes(Refs{})) // on the response

	container.Add(ws)
}

func (u RefResource) findRefs(request *restful.Request, response *restful.Response) {
	h, err := hex.DecodeString(request.PathParameter("hash"))
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Hash could not be found.")
		return
	}
	err = u.ref.GetRawRef(h, func(v []byte) error {
		r := Refs{
			request.PathParameter("hash"),
			hex.EncodeToString(v),
		}
		fmt.Println(v)
		response.WriteEntity(r)
		return nil
	})
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: User could not be found.")
		return
	}
	return
}

func main() {

	wsContainer := restful.NewContainer()
	rdb, err := badgerdb.NewDB("../temp" + "/refs")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rdb.Close()
	u := NewRefResource(rdb)
	u.Register(wsContainer)

	config := swagger.Config{
		WebServices:    wsContainer.RegisteredWebServices(), // you control what services are visible
		WebServicesUrl: "http://localhost:8080",
		ApiPath:        "/apidocs.json",

		// Optionally, specifiy where the UI is located
		SwaggerPath:     "/apidocs/",
		SwaggerFilePath: "/Users/emicklei/xProjects/swagger-ui/dist"}
	swagger.RegisterSwaggerService(config, wsContainer)

	log.Printf("start listening on localhost:8080")
	server := &http.Server{Addr: ":8080", Handler: wsContainer}
	log.Fatal(server.ListenAndServe())
}
