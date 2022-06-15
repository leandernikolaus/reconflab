# Reconfiguration of replicated systems

This repository contains exercises for my lecture on reconfigurable replicated systems.
Exercises are small programming tasks. Examples are build using the [Gorums](https://github.com/relab/gorums/blob/master/doc/user-guide.md) RPC framework.

* [storage](./storage/) contains some small tasks to get to know the example application and the Gorums framework.
* [reconfstorage](./reconfstorage/) contains tasks to implement reconfiguration mechanisms for the same application.

## Disclaimer

It is my hope that these exercises will eventually be self-explanatory. 
However, in their current state the design may be difficult to understand without the accompanying lecture covering both algorithms for reconfiguration and the main workings of the Gorum RPC framework.


## Getting started

You should have a recent version of Go installed. Installation instructions can be found [here](https://go.dev/doc/install).

To get started, clone this repository and open the [storage](./storage/) folder in your favorite Golang editor.
Being done with the exercises their, do the same with the [reconfstorage](./reconfstorage/) folder.