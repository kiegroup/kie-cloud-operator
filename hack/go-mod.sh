#!/bin/sh

echo Reset vendor diectory

go mod vendor
go mod verify