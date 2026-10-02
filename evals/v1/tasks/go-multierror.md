# go-multierror task

At `hashicorp/go-multierror` commit `6d4d48630db25c3c83fa83ecd41dd8438b82963c`, add a documented `(*Error).ErrorsSnapshot() []error` method. It must return a shallow copy in original order, return nil for a nil receiver, and avoid exposing the receiver's slice storage. Add focused tests, keep the existing APIs compatible, and run the package tests. Do not commit or push.
