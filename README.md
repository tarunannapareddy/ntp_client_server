# NTP Client and Server Implementation

This repo is an implementation of the NTP Client Server for time synchronization. 
## Overview

The NTP protocol is designed to synchronize the clocks of computer systems over packet-switched, variable-latency data networks. The implementation in this repository is a scaled-down version of the NTP protocol, as per the requirements of the assignment.

## Implementation

### Server
The server implementation listens for incoming NTP client requests on the specified port. When a request is received, the server processes the request, updates the relevant fields in the NTP packet, and sends a response back to the client.

### Client

The client implementation sends NTP requests to the server in bursts of eight packets every four minutes for one hour. The delays and offsets are calculated based on the timestamps in the NTP packets. The measured delays and offsets are stored in slices and plotted using the `gonum/plot` package. 
The resulting plots are saved as PNG files in the current directory.


