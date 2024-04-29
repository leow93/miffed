
const client = new WebSocket('ws://localhost:8080/socket')


client.onopen = (ev) => {
    console.log('Connection established')
}

client.onclose = (ev) => {
    console.log('Connection closed')
}

client.onerror = (ev) => {
    console.log('Error:', ev)
}

client.onmessage = (ev) => {
    console.log('Message:', ev.data)
}
client.send('Hello, server!')