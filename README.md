# go-httpio
A minimal utility that makes it trivial to parse and render HTTP by using input and output structs

sits between your router and your business logic and takes care of boring stuff like:

- bring your own router
- freedom in rendering output (e.g: templates instead of json)
- allow you to take responsibility of errors: customize rendering
- generated clients
- content negotiation build in
- pluggable encoders and decoders
- plugin form decoding
- custom rendering into parsing logic

- TODO: plugin validation
- TODO: context value injection using tags
