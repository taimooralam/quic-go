This repository contains both the prototype and the experiment for the thesis.

The experiment contains two parts `cli.go` which is a QUIC client in GO. 

The experiment measures the latency over QUIC of packet transmission in mininet with a base-latency of 30ms and packet loss of 1%. It sends the packets from the QUIC client `cli.go` to a QUIC server `serv.go` and measure the TT of one-way target latency of the packets and how quickly they arrrive against their sequence numbers. The data is then stored in a file called `Log.txt` (whose path can be set in `serv.go`) and then plotted with plotly to give the latencies of the packets.

In the the prototype consists of two parts, a QUIC server called `pro_serv.go` which acts as the operator end-point for reading point-cloud QUIC data over the network and a QUIC client called `pro_client.go` which sends compressed point-cloud data to the QUIC server.

The point-cloud compression code called `pcd_read` at the vehicle end-point and the point-cloud decompression code called `pcd_write` at the operator end-point are both placed in the master thesis repository at https://github.com/taimooralam/master-thesis/tree/master/prototype/point-cloud-streaming. At the vehicle end-point `pcd_read` reads the raw point-cloud data from a file and sends it via named pipes to `pro_cli.go` which then transmits it over the network in QUIC to `pro_serv.go` at the operator end-point. `pro_serv.go` then transmits that data to `pcd_write` at the operator end-point through named-pipes. The data is then decompressed using Octree Decompression and saved in `data` directory on the operator end-point.

Whenever running the experiment please check the IP address and port numbers for QUIC transmission in all the four files: `cli.go`, `serv.go`, `pro_cli.go` and `pro_serv.go`. IN `serv.go` also check the path of the log file that will store the contents of the latency measurements. This `Log.txt` file can then be used to plot the latency graph of the latency measurements in the experiment using `graph.py` file in the master-thesis repository. 
