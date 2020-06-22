# Third Part Imports

This folder contains third party dependencies for protobufs. Because go modules will not deal with `.proto` dependencies, we are forced to manually copy these here. If we do have to update them, it would required a manual action. However, in practice we are unlikely to ever need to. 

The generated `pb.go` files will use the dep from go.mod, so on that end we will always stay up to date. In the `.proto` file, we only depend on TypeMeta, ObjectMeta, ListMeta, and LabelSelector. These are unlikely to change. If they did, it would cause issues because the `pb.go` file will still be updated; we will only have issues if those types change and service-apis itself needs to use the new fields, in which case we will manually update the .proto files here.
