# E2E

[![GoDoc](https://godoc.org/github.com/arussellsaw/e2e?status.svg)](https://godoc.org/github.com/arussellsaw/e2e)

a package to help you run end to end tests on your services and infrastructure. This package aims to allow users to write tests in the same they would with the testing package, and have them run continuously in production, to alert on regressions and issues not picked up by unit and integration testing. This package is still in development, and will be liable to breaking changes.

![web ui](https://i.imgur.com/xA9fxJ9.jpg)

## TODO
* Notifier interface
* slack notifier
* webhook notifier
* email notifier
* pagerduty notifier
* influx notifier
* prometheus notifier
* debounce/flapping test handling
* fuzzing utilities
* more advanced web ui

## Done
* satisfy testing.TB interface
* basic web ui
* scheduling of test runs
