#!/bin/sh
flyctl deploy --build-arg "VERSION=$(git describe --tag)"
