#!/bin/bash

ginkgo -cover && go tool cover -html=hellosign.coverprofile

