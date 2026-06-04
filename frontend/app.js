const chat = document.getElementById('chat');
const inputForm = document.getElementById('inputForm');
const queryInput = document.getElementById('queryInput');
const sendBtn = document.getElementById('sendBtn');
const sourcesPanel = document.getElementById('sourcesPanel');
const sourcesList = document.getElementById('sourcesList');
const sourcesCount = document.getElementById('sourcesCount');

let loading = false;

inputForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const query = queryInput.value.trim();
    if (!query || loading) return;

    addMessage(query, 'user');
    queryInput.value = '';
    showTyping();

    loading = true;
    sendBtn.disabled = true;

    try {
        const res = await fetch('/api/search', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ query }),
        });

        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Request failed');
        }

        const data = await res.json();
        removeTyping();
        addMessage(data.answer, 'assistant');
        showSources(data.sources);
    } catch (err) {
        removeTyping();
        addMessage('Error: ' + err.message, 'assistant error');
        showSources([]);
    } finally {
        loading = false;
        sendBtn.disabled = false;
        queryInput.focus();
    }
});

function addMessage(text, role) {
    const div = document.createElement('div');
    div.className = `message ${role}`;

    const avatar = document.createElement('div');
    avatar.className = 'avatar';
    avatar.textContent = role === 'user' ? 'U' : 'R';

    const bubble = document.createElement('div');
    bubble.className = 'bubble';

    if (role === 'assistant') {
        // Render thinking/reasoning if present (between  and )
        const thinkMatch = text.match(/<think>([\s\S]*?)<\/think>/);
        if (thinkMatch) {
            const thinkDiv = document.createElement('div');
            thinkDiv.className = 'thinking';
            thinkDiv.textContent = 'Thinking: ' + thinkMatch[1].trim();
            bubble.appendChild(thinkDiv);

            const answer = text.replace(/<think>[\s\S]*?<\/think>\s*/, '').trim();
            if (answer) {
                bubble.appendChild(document.createTextNode(answer));
            }
        } else {
            bubble.textContent = text;
        }
    } else {
        bubble.textContent = text;
    }

    div.appendChild(avatar);
    div.appendChild(bubble);
    chat.appendChild(div);
    chat.scrollTop = chat.scrollHeight;
}

function showTyping() {
    const div = document.createElement('div');
    div.className = 'message assistant';
    div.id = 'typingIndicator';

    const avatar = document.createElement('div');
    avatar.className = 'avatar';
    avatar.textContent = 'R';

    const bubble = document.createElement('div');
    bubble.className = 'bubble';
    bubble.innerHTML = '<div class="typing-indicator"><span></span><span></span><span></span></div>';

    div.appendChild(avatar);
    div.appendChild(bubble);
    chat.appendChild(div);
    chat.scrollTop = chat.scrollHeight;
}

function removeTyping() {
    const el = document.getElementById('typingIndicator');
    if (el) el.remove();
}

function showSources(sources) {
    sourcesList.innerHTML = '';
    sourcesCount.textContent = sources.length;

    if (sources.length === 0) {
        sourcesPanel.classList.remove('open');
        return;
    }

    sources.forEach((src) => {
        const tag = document.createElement('span');
        tag.className = 'source-tag';
        tag.textContent = src;
        sourcesList.appendChild(tag);
    });

    sourcesPanel.classList.add('open');
}

queryInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        inputForm.dispatchEvent(new Event('submit'));
    }
});
