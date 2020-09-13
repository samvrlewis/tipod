# tipod

tipod (Tunelled IP over DNS) allows IP (v4) traffic to be tunnelled over DNS packets. It's useful in scenarios where normal traffic is blocked but DNS queries are permitted.

The current version of the software is very rough around the edges and is primarily being written as a learning excercise. As it stands, it probably should not be used for anything serious, but if you are interested in using it for something serious feel free to contribute towards making it more robust! In the meantime, [iodine](https://github.com/yarrick/iodine) is project (written in C) created for a similiar purpose which may be suitable.

## Todo

- [ ] Authentication
- [ ] Allowing multiple clients on the server
- [ ] Documentation
- [ ] ipv6 support
- [ ] Benchmarking
- [ ] Optimising for bandwidth
- [ ] Multi platform support (tipod only currently supports Linux)
