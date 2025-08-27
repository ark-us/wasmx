#!/usr/bin/make -f

tidy:
	find . -name go.mod -execdir go mod tidy \;
