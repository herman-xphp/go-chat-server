let ws;
let username = '';
let onlineUsers = new Set();

function joinChat() {
    const usernameInput = document.getElementById('usernameInput');
    username = usernameInput.value.trim();

    if (!username) {
        alert('Please enter a username');
        return;
    }

    document.getElementById('loginContainer').style.display = 'none';
    document.getElementById('chatContainer').style.display = 'grid';
    document.getElementById('currentUser').textContent = username;

    connectWebSocket();
}

function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws?username=${encodeURIComponent(username)}`;

    ws = new WebSocket(wsUrl);

    ws.onopen = function() {
        console.log('Connected to chat server');
        addSystemMessage('Connected to chat server');
    };

    ws.onmessage = function(event) {
        const message = JSON.parse(event.data);
        handleMessage(message);
    };

    ws.onerror = function(error) {
        console.error('WebSocket error:', error);
        addSystemMessage('Connection error');
    };

    ws.onclose = function() {
        console.log('Disconnected from chat server');
        addSystemMessage('Disconnected from server');
    };
}

function handleMessage(message) {
    switch (message.type) {
        case 'join':
            addSystemMessage(message.content);
            if (message.from !== username) {
                onlineUsers.add(message.from);
                updateUsersList();
            }
            break;

        case 'leave':
            addSystemMessage(message.content);
            onlineUsers.delete(message.from);
            updateUsersList();
            break;

        case 'broadcast':
            addChatMessage(message);
            break;

        case 'private':
            addChatMessage(message, true);
            break;

        case 'user_list':
            // Handle user list update
            break;
    }
}

function sendMessage() {
    const input = document.getElementById('messageInput');
    const content = input.value.trim();

    if (!content || !ws) {
        return;
    }

    const message = {
        type: 'broadcast',
        content: content,
        timestamp: new Date().toISOString()
    };

    ws.send(JSON.stringify(message));
    input.value = '';
}

function handleKeyPress(event) {
    if (event.key === 'Enter') {
        sendMessage();
    }
}

function addChatMessage(message, isPrivate = false) {
    const messagesDiv = document.getElementById('messages');
    const messageDiv = document.createElement('div');
    
    const isOwnMessage = message.from === username;
    messageDiv.className = `message ${isOwnMessage ? 'own' : 'broadcast'}`;

    const senderDiv = document.createElement('div');
    senderDiv.className = 'message-sender';
    senderDiv.textContent = isOwnMessage ? 'You' : message.from;

    const contentDiv = document.createElement('div');
    contentDiv.className = 'message-content';
    contentDiv.textContent = message.content;

    const timeDiv = document.createElement('div');
    timeDiv.className = 'message-time';
    timeDiv.textContent = new Date(message.timestamp).toLocaleTimeString();

    messageDiv.appendChild(senderDiv);
    messageDiv.appendChild(contentDiv);
    messageDiv.appendChild(timeDiv);

    messagesDiv.appendChild(messageDiv);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

function addSystemMessage(content) {
    const messagesDiv = document.getElementById('messages');
    const messageDiv = document.createElement('div');
    messageDiv.className = 'message system';
    messageDiv.textContent = content;

    messagesDiv.appendChild(messageDiv);
    messagesDiv.scrollTop = messagesDiv.scrollHeight;
}

function updateUsersList() {
    const usersList = document.getElementById('usersList');
    usersList.innerHTML = '';

    // Add self
    const selfLi = document.createElement('li');
    selfLi.textContent = `${username} (You)`;
    usersList.appendChild(selfLi);

    // Add others
    onlineUsers.forEach(user => {
        if (user !== username) {
            const li = document.createElement('li');
            li.textContent = user;
            usersList.appendChild(li);
        }
    });
}

function leaveChat() {
    if (ws) {
        ws.close();
    }

    document.getElementById('chatContainer').style.display = 'none';
    document.getElementById('loginContainer').style.display = 'flex';
    document.getElementById('usernameInput').value = '';
    document.getElementById('messages').innerHTML = '';
    onlineUsers.clear();
}

// Initialize
updateUsersList();
