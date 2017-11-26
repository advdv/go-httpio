# go-httpio
Go is a fantastic language for developing web services and it is said that the standard library is
powerfull enough (with one of the many routers) to achieve this. Other claim that this is not good enough when
you want to hit the ground running and setup a new project up quickly as you'll be writing a lot of boilerplate code.
The latter has let to the development of more holistic web frameworks that take care of this but these
come with a lot more then you might need and can be very opinionated.

This project aims to find a middleground by only focussing on removing the boilerplate code that is
required for parsing HTTP requests and writing back a (error) response. As such it sits between your router and  
and the business logic you're trying to focus development effort on.


## Features
- Build fully on the standard library without any mandatory external dependencies
- Bring your own router, any multiplexer in the ecosystem that can route to an http.HandlerFunc is usable
- Takes care of decoding request bodies and encoding to responses using a flexible stack of encoders and the Content-Type header
- Uses these encoding stacks to do content negotiation based on the Accept header
- Provide a central error handling mechanism for logging or providing specific user feedback
- Comes with a http client that be used to write easily write client side code
- Optionally allows form value decoding and encoding using third-party libraries
- Optionally allows parsed request bodies to be validated using third-party libraries
- Optionally allows full rendering customization, for example to support template rendering.

## An Example
Consider the development of a web service that manages accounts. You would to like to focus on the logic
of creating accounts not on the work that is required to turn http.Requests into its input or turning
create account output into responses. Ideally you would like to isolate this like so:

```Go
type MyAccountService struct{}

type CreateAccountInput struct{
  Name string
}

type CreateAccountOutput struct{
  ID string
  Name string
}

func (ctrl *MyAccountCtrl) CreateAccount(ctx context.Context, input *CreateAccountInput) (*CreateAccountOutput, error) {
  //you would like to focus on what happens here
}
```

Instead of having to write your own decoding and encoding logic for each input or output struct this library
simply allows you do the following:

```Go
r := mux.NewRouter() //bring your own router, for example github.com/gorilla/mux
svc := &MyAccountService{} //this holds your account creation implementation

ctrl := httpio.NewCtrl(
  &encoding.JSON{}, //enables JSON decoding for inputs and JSON encoding for outputs
)

ctrl.SetValidator(/* choose an validator from the eco system */)
ctrl.SetErrorHandler(/* allow error handling to be customized */)

//and voila, your request handlers can now look like this. Notice you don't have to write any
//logic for decoding the CreateAccountInput or encoding its output.
r.HandleFunc("/accounts/create", func(w http.ResponseWriter, r *http.Request) {
  input := &CreateAccountInput{}
  if render, ok := ctrl.Handle(w, r, input); ok {
    render(svc.CreateAccount(r.Context(), input))
  }
})

//now serve the router and your good to go
log.Fatal(http.ListenAndServe(":8080", r))
```

## Recipes
Although the library designed to be flexible and serve different needs for
different web applications. Much of these are still need to written but you
can take a look at the `examples/sink` code for most of it.

- Using `*template.Templates` to render Outputs: WIP
- Using the `github.com/go-playground/validator` validator: WIP
- Allow inputs to be decoded from from submissions and query parameters: WIP
- Handle certain (user) errors differently: WIP
- Customize response status code: WIP
- Disable the 'X-Has-Handling-Error' header: WIP
- Using the client with application specific errors: WIP

## Future Ideas

### Adding ctx values to implementation inputs
Another part that can be cumbersome in setting up for your business logic is fetching values from
the request context that are set by certain middleware, e.g: session, request ids etc. It would be
cool if the struct tags of input could indicate that their value should be fetched from the request
context:

```
type MyInput {
  Session *Session `json:"-" form:"-" ctx:"MySession"`
}
```  

_httpio_ then would take care of injecting these into the input struct before handing it off to the
business logic. Main problem is that it is recommended to use package specific type values for context
keys, so a annotation as the one above would not be possible. If you have any ideas for this let me know.

### Adding header values to the implementation input
Similarly some implementation functions might want to retrieve a specific header. Middleware would allow
this for a set of endpoints but sometimes the client might send over a specific header.
