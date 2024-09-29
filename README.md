
# 🌊 GRE Flooder
### Version: v1
This Go-based packet flooder utilizes raw sockets to send packets to a specified target IP address.

## 🚀 Overview
The gre flooder is designed to send a high volume of packets to a target server, which can be useful for testing network robustness and stress-testing server performance.

## ⚙️ Features
- **Custom Packet Size:** The flooder can send packets up to 65535 bytes in size.
- **Protocol Support:** Utilizes the GRE protocol (Protocol Number: 47) for encapsulating packets.
- **Rate Limiting:** Configurable packets-per-second (PPS) limiter to control the flood rate.
- **Multi-threading:** Supports multiple threads to maximize flooding efficiency.
- **Dynamic Sleep Adjustment:** Automatically adjusts sleep time between packets based on the defined PPS limit.

## 📦 How It Works
1. **Initialization:** The application takes command-line arguments for the target IP address, number of threads, packets-per-second limit, and duration of the attack.
2. **Socket Creation:** Raw sockets are established for sending packets.
3. **Packet Construction:** Each packet consists of an IP header and a TCP header followed by a payload.
4. **Flooding Logic:** Continuously sends packets until the specified duration expires, respecting the PPS limit.

## Usage
To run the application, use the following command:
```
go run <target IP> <number of threads> <pps limiter, -1 for no limit> <time in seconds>
```

### Example
```bash
go run main.go 192.168.1.1 10 100 60
```
This command will flood the target IP `192.168.1.1` using 10 threads, allowing up to 100 packets per second for 60 seconds.

## Important Note
**Use responsibly.** This tool is intended for educational and testing purposes only. Ensure you have permission to test the target network.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
