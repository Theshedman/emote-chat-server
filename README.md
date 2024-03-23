# Emote Chat System

...a scalable real-time chat application implemented in GoLang

## How to Set up and Run the server
#### Prerequisite
1. Make sure that Go is installed and configured in your system
2. A MongoDB instance for the database
3. A RabbitMQ instance for managing real-time chat at scale
4. Clone the project from `https://github.com/Theshedman/emote-chat-server.git`
5. change directory to the project's folder: `cd emote-chat-server`

### Using docker
1. Ensure that you have docker installed and running in your system
2. Build the docker image using the command:
```bash
docker -t <tagName> .
```
`<tagName>` can be any name you would like to tag the docker image with. Also, pay attention to the period (.) at the end of the build command

3. Run the built image with this command:
```bash
docker run -d \
-e DB_URL=<DB_URL> \
-e DB_NAME=<DB_NAME> \
-e RABBITMQ_URL=<RMQ_URL> \
-e JWT_SECRET=<JWT_SECRET> \
-e SERVER_PORT=<SERVER_PORT> \
<tagName>
```
Please note that you should replace the `<tagName>` with the very name you used during the docker image build process. Also, provide the values for the environment variables

### Using Local Server
1. Add `.env` file on the project root path
2. Add these variables to the file:
```text
DB_URL=<DB_URL>
DB_NAME=<DB_NAME>
RABBITMQ_URL=<RABBITMQ_URL>
JWT_SECRET=<JWT_SECRET>
SERVER_PORT=<SERVER_PORT>
```
3. Run this command to start the server locally:
```bash
go run main.go
```