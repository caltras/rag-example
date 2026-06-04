const chat = document.getElementById('chat');
const inputForm = document.getElementById('inputForm');
const queryInput = document.getElementById('queryInput');
const sendBtn = document.getElementById('sendBtn');
const sourcesPanel = document.getElementById('sourcesPanel');
const sourcesList = document.getElementById('sourcesList');
const sourcesCount = document.getElementById('sourcesCount');
const modeSwitch = document.getElementById('modeSwitch');

const tabs = document.querySelectorAll('.tab');
const articlesSubtitle = document.querySelector('.articles-subtitle');
const travelSubtitle = document.querySelector('.travel-subtitle');
const welcomeArticles = document.getElementById('welcome-articles');
const welcomeTravel = document.getElementById('welcome-travel');

let loading = false;
let currentMode = 'articles';
let sessionId = localStorage.getItem('ragSessionId') || crypto.randomUUID();
let travelSessionId = localStorage.getItem('ragTravelSessionId') || crypto.randomUUID();

tabs.forEach(tab => {
    tab.addEventListener('click', () => {
        tabs.forEach(t => t.classList.remove('active'));
        tab.classList.add('active');
        currentMode = tab.dataset.tab;

        if (currentMode === 'articles') {
            articlesSubtitle.style.display = '';
            travelSubtitle.style.display = 'none';
            welcomeArticles.style.display = '';
            welcomeTravel.style.display = 'none';
            queryInput.placeholder = 'Ask a question...';
        } else {
            articlesSubtitle.style.display = 'none';
            travelSubtitle.style.display = '';
            welcomeArticles.style.display = 'none';
            welcomeTravel.style.display = '';
            queryInput.placeholder = 'Plan your trip...';
        }
    });
});

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
        const endpoint = currentMode === 'travel' ? '/api/travel/search' : '/api/search';
        const mode = modeSwitch.checked ? 'data' : 'chat';
        const body = currentMode === 'travel'
            ? JSON.stringify({ query, session_id: travelSessionId, mode })
            : JSON.stringify({ query, session_id: sessionId, mode });

        const res = await fetch(endpoint, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body,
        });

        if (!res.ok) {
            const text = await res.text();
            throw new Error(text || 'Request failed');
        }

        const data = await res.json();
        removeTyping();
        addTimedMessage(data.answer, 'assistant', data, data.mode === 'data');

        if (currentMode === 'travel') {
            localStorage.setItem('ragTravelSessionId', data.session_id || travelSessionId);
            showTravelSources(data);
        } else {
            localStorage.setItem('ragSessionId', data.session_id || sessionId);
            showSources(data.sources || []);
        }
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
    addTimedMessage(text, role, null, false);
}

function addTimedMessage(text, role, timingData, isDataMode) {
    const div = document.createElement('div');
    div.className = `message ${role}`;
    if (isDataMode) div.classList.add('data-mode');

    const avatar = document.createElement('div');
    avatar.className = 'avatar';
    avatar.textContent = role === 'user' ? 'U' : (currentMode === 'travel' ? 'T' : 'R');

    const bubble = document.createElement('div');
    bubble.className = 'bubble';

    if (role === 'assistant') {
        const thinkMatch = text.match(/<think>([\s\S]*?)<\/think>/);
        let body = text;
        if (thinkMatch) {
            const thinkDiv = document.createElement('div');
            thinkDiv.className = 'thinking';
            thinkDiv.textContent = 'Thinking: ' + thinkMatch[1].trim();
            bubble.appendChild(thinkDiv);
            body = text.replace(/<think>[\s\S]*?<\/think>\s*/, '').trim();
        }
        if (body) {
            const content = document.createElement('div');
            content.className = 'markdown';
            if (isDataMode) {
                const badge = document.createElement('div');
                badge.className = 'data-badge';
                badge.textContent = 'Raw Data';
                content.appendChild(badge);
            }
            content.innerHTML += marked.parse(body);
            bubble.appendChild(content);
        }

        if (timingData && timingData.timing_total_ms != null) {
            const timing = document.createElement('div');
            timing.className = 'timing';
            const e = timingData.timing_embed_ms;
            const s = timingData.timing_search_ms;
            const c = timingData.timing_chat_ms;
            const t = timingData.timing_total_ms;
            timing.textContent = `⏱ embed ${e}ms · search ${s}ms · chat ${(c/1000).toFixed(1)}s · total ${(t/1000).toFixed(1)}s`;
            bubble.appendChild(timing);
        }
    } else {
        bubble.textContent = text;
    }

    div.appendChild(avatar);
    div.appendChild(bubble);
    chat.appendChild(div);
    chat.scrollTop = chat.scrollHeight;
}

function showTravelSources(data) {
    sourcesList.innerHTML = '';
    const allSources = [];

    if (data.flights && data.flights.length > 0) {
        data.flights.forEach(f => {
            allSources.push(`✈ ${f.airline} ${f.flight_number}: $${f.price}`);
        });
    }
    if (data.hotels && data.hotels.length > 0) {
        data.hotels.forEach(h => {
            allSources.push(`🏨 ${h.name}: $${h.price_per_night}/night`);
        });
    }
    if (data.destinations && data.destinations.length > 0) {
        data.destinations.forEach(d => {
            allSources.push(`📍 ${d.title}`);
        });
    }

    sourcesCount.textContent = allSources.length;

    if (allSources.length === 0) {
        sourcesPanel.classList.remove('open');
        return;
    }

    allSources.forEach(src => {
        const tag = document.createElement('span');
        tag.className = 'source-tag';
        tag.textContent = src;
        sourcesList.appendChild(tag);
    });

    sourcesPanel.classList.add('open');
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

function showTyping() {
    const div = document.createElement('div');
    div.className = 'message assistant';
    div.id = 'typingIndicator';

    const avatar = document.createElement('div');
    avatar.className = 'avatar';
    avatar.textContent = currentMode === 'travel' ? 'T' : 'R';

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

queryInput.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        inputForm.dispatchEvent(new Event('submit'));
    }
});
