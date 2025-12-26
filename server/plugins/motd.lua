print("MotD Plugin Started")
on("client_connect", function(client_id)
    print('client connected! ')
    print(client_id)
    send_packet(client_id, {
        type="chat_message",
        data={
            sender={
                id="server"
            }
        }
    })
end)