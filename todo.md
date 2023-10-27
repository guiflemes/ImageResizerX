## Features

* Add GRPC entrypoint
* Add Unit Tests
* Create a system to delete image after the download 
    - clean method [V]
    - create queue to delete it []
* Inject dimension by client

## Improvements

* Add host and port to ListenAndServe
* Replace logs like json to default pattern
* Add log to any router, when the request hits it, this should show log up

## Fix

* Fix Download (Its on front-end)