## Factor

Factor is a relay which can be used to keep a number of EL clients supplied with recent-head
information, based off one CL client. 

The relay currently only supports one mode -- active mode. 

**OBS** Factor is not a secure way to manage an EL node. 

### Active mode

In active mode, the relay fetches information from the CL node, and then passes it down to the 
EL nodes that are configured. 

### Passive mode

In passive mode, the relay functions like an EL node -- and the CL pushes changes to it. It then 
relays the data to other nodes, but uses the primary EL to return responses. 

## Configuration

Factor can handle jwt and custom headers. See `conf.toml.sample` for an idea of how to configure it. 


## Docker 

Should be available at [docker hub](https://hub.docker.com/r/holiman/factor)