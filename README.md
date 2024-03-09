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

Delay Values when the Client, Server are on the same Network

<img width="459" alt="Screenshot 2024-03-08 at 10 58 53 PM" src="https://github.com/tarunannapareddy/ntp_client_server/assets/19953916/94f18d6f-ca3e-4e8a-9481-d0554880467d">

Delay values with NTP Servers

<img width="396" alt="Screenshot 2024-03-08 at 10 59 57 PM" src="https://github.com/tarunannapareddy/ntp_client_server/assets/19953916/51d53682-b97b-46f6-b4a2-67d59f09231d">

Delay values with Client/Server on different regions of the cloud

<img width="396" alt="Screenshot 2024-03-08 at 11 00 26 PM" src="https://github.com/tarunannapareddy/ntp_client_server/assets/19953916/fec63000-e5e6-43cc-8db8-9eab41728443">
