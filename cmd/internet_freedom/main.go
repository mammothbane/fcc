package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/schema"
	"github.com/mammothbane/fcc"
	"gopkg.in/go-playground/validator.v9"
)

var (
	assets     = http.FileServer(http.Dir("assets"))
	dec        = schema.NewDecoder()
	proceeding *fcc.Proc
	validate      = validator.New()
	tmpl              = template.Must(template.ParseGlob("tmpl/*"))
)

type formContent struct {
	Name    string `form:"name" validate:"required"`
	Email   string `form:"email" validate:"required"`
	Address string `form:"address" validate:"required"`
	City    string `form:"city" validate:"required"`
	Zip     string `form:"zip" validate:"required"`
	State   string `form:"state" validate:"required"`
	Comment string `form:"comment" validate:"required"`
}

func (f formContent) toFilingInfo() *fcc.FilingInfo {
	return &fcc.FilingInfo{
		Named: fcc.Named{Name: f.Name},
		Address: fcc.Address{
			AddressFirstLine: f.Address,
			City:             f.City,
			ZipCode:          f.Zip,
			State:            f.State,
		},
		Text:  f.Comment,
		Email: f.Email,
	}
}

const (
	PROCEEDING = "17-108"
	COOKIE_ID  = "clientid"
)

func handleSubmit(c *gin.Context) {
	//filing, err := fcc.Status("20170509864119531")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//fmt.Println(filing)

	log.Println("got submit")

	bail := func(err error, failcode int, usermsg string, msg string, args ...interface{}) bool {

		errMap := map[string]interface{}{
			"message": usermsg,
			"code":    failcode,
		}

		if err != nil {
			log.Printf(msg, append(args, err)...)
			c.HTML(failcode, "error.html", errMap)
			return true
		}
		return false
	}

	var content formContent
	if c.Bind(&content) != nil {
		return
	}
	info := content.toFilingInfo()

	bck := &backoff{
		factor:  2,
		current: 10 * time.Millisecond,
		max:     10 * time.Second,
	}

	var conf *fcc.FilingConfirmation

	log.Println("submitting filing info...")

	err := bck.do(func() error {
		inner, err := info.Submit(proceeding)
		if err == nil {
			conf = inner
		} else {
			log.Printf("Error: '%v'. Retrying.", err)
		}

		return err
	})

	if bail(err, http.StatusInternalServerError, "Sorry, the FCC server wouldn't accept the filing.", "Failed submitting filing info: %v") {
		return
	}

	log.Println("Successfully submitted! Confirmation:", conf.Confirmation)
}

func main() {

	bck := &backoff{
		factor:  2,
		current: 10 * time.Millisecond,
		max:     10 * time.Second,
	}

	log.Println("retrieving proceeding...")
	err := bck.do(func() error {
		inner, err := fcc.Proceeding(PROCEEDING)
		if err == nil {
			proceeding = inner
		} else {
			log.Printf("Error: '%v'. Retrying.", err)
		}

		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println("got proceeding", PROCEEDING)

	engine := gin.Default()

	g1 := engine.Group("/")
	g1.StaticFS("/", http.Dir("assets"))
	g1.POST("/submit", handleSubmit)
	g1.Use(func(c *gin.Context) {
		cookie, err := c.Request.Cookie(COOKIE_ID)
		if err != nil || cookie == nil {
			c.Next()
		} else {
			c.Redirect(http.StatusTemporaryRedirect, "/check")
		}

	})

	engine.LoadHTMLGlob("tmpl/*")
	log.Fatal(engine.Run(":8080"))
}

type backoff struct {
	factor  float64
	current time.Duration
	max     time.Duration
}

func (b *backoff) do(op func() error) error {
	for {
		err := op()
		if err == nil {
			return err
		}

		b.current = time.Duration(float64(b.current) * b.factor)
		if b.current > b.max {
			return err
		}

		time.Sleep(b.current)
	}
}
