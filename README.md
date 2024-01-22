# Go-Generate

## NOTE: STILL IN DEVELOPMENT NO MANY MODES

Currently has only the capability to extract function from a package. Lets say something like this

Package A has 10 functions

You want to extract 2 functions to a different package

Prepare a Directory B with interface and the 2 functions you want to extract
Run this executable with the args
```
go-generate <directory to source package> <direcetory to Destination (B)> <New Package name>
```