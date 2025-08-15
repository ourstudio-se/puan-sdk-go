# PUAN SDK GO

SDK for ILP modelling and matric creation.

## Pre-requisites

### Setting up Rust GLPK API

1. Clone source git repository [here](https://github.com/ourstudio-se/rust-glpk-api).
2. Stand at the root of the `rust-glpk-api` repository locally.
3. Build image: `docker build -t glpk-api .`

## Running

### Running Rust GLPK API

Make sure the images is setup (see **Setting up Rust GLPK API** under Pre-requsites)

Run `make glpk`. This will start a detached glpk api Docker container. The default port of the api is `9000`.
