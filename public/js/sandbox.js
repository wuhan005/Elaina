(() => {
    const config = window.ELAINA;
    const text = config.text;
    const $ = (id) => document.getElementById(id);

    let lang = config.language;
    // Code per language; an embedding page can preset it through postMessage.
    let languagePlaceholder = {};
    let running = false;

    const editorMode = (language) => (language === 'c' ? 'text/x-csrc' : language);

    const editor = CodeMirror.fromTextArea($('code'), {
        lineNumbers: true,
        mode: editorMode(lang),
        theme: 'elaina',
        indentUnit: 4,
        tabSize: 4,
        readOnly: !config.editable,
        extraKeys: {
            'Ctrl-Enter': () => run(),
            'Cmd-Enter': () => run(),
        },
    });
    editor.setSize('100%', '100%');

    // Theme follows the system unless ?theme= pinned it.
    const themeParam = new URLSearchParams(location.search).get('theme');
    if (themeParam !== 'light' && themeParam !== 'dark') {
        matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (event) => {
            document.documentElement.dataset.theme = event.matches ? 'dark' : 'light';
        });
    }

    const isMac = /Mac|iPhone|iPad/.test(navigator.platform);
    $('run-shortcut').textContent = isMac ? '⌘↵' : 'Ctrl ↵';

    // Output rendering.
    const output = $('output');
    const status = $('status');
    const duration = $('duration');
    const clearButton = $('clear');

    function formatDuration(nanoseconds) {
        const ms = nanoseconds / 1e6;
        if (ms < 1) return '<1 ms';
        if (ms < 1000) return `${Math.round(ms)} ms`;
        return `${(ms / 1000).toFixed(2)} s`;
    }

    function element(tag, className, content) {
        const node = document.createElement(tag);
        if (className) node.className = className;
        if (content !== undefined) node.textContent = content;
        return node;
    }

    function showStatus(kind, label) {
        status.hidden = false;
        status.className = `status status-${kind}`;
        status.textContent = label;
    }

    function reset() {
        output.replaceChildren(element('p', 'placeholder', text.idle));
        status.hidden = true;
        duration.textContent = '';
        clearButton.hidden = true;
    }

    function renderFailure(message) {
        output.replaceChildren();
        const box = element('div', 'error-box');
        box.append(element('span', 'error-label', text.error), element('pre', null, message));
        output.append(box);
        showStatus('failed', text.request_failed);
        duration.textContent = '';
        clearButton.hidden = false;
    }

    function renderResult(result) {
        const {stdout = '', stderr = '', error = ''} = result;
        output.replaceChildren();

        if (error) {
            const box = element('div', 'error-box');
            box.append(element('span', 'error-label', text.error), element('pre', null, error));
            output.append(box);
        }
        if (stdout) {
            output.append(element('pre', 'stdout', stdout));
        }
        if (stderr) {
            const section = element('div', 'stderr');
            section.append(element('span', 'stream-label', text.stderr), element('pre', null, stderr));
            output.append(section);
        }
        if (!error && !stdout && !stderr) {
            output.append(element('p', 'placeholder', text.no_output));
        }

        if (!error) showStatus('success', text.success);
        else if (error === 'execution timed out') showStatus('timeout', text.timed_out);
        else showStatus('failed', text.failed);
        duration.textContent = result.duration ? formatDuration(result.duration) : '';
        clearButton.hidden = false;
    }

    function setRunning(value) {
        running = value;
        const button = $('run');
        button.disabled = value;
        button.classList.toggle('is-running', value);
        $('run-label').textContent = value ? text.running : text.run;
        output.classList.toggle('is-stale', value);
    }

    async function run() {
        if (running) return;
        setRunning(true);
        try {
            const response = await fetch(`${location.pathname.replace(/\/$/, '')}/execute`, {
                method: 'POST',
                headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                body: new URLSearchParams({lang, code: editor.getValue()}),
            });
            const body = await response.json().catch(() => null);
            if (!response.ok || !body?.data) {
                renderFailure(body?.msg || `${text.request_failed} (${response.status})`);
                return;
            }
            renderResult(body.data.result);
        } catch {
            renderFailure(text.request_failed);
        } finally {
            setRunning(false);
        }
    }

    $('run').addEventListener('click', run);
    clearButton.addEventListener('click', reset);

    // Language switch keeps the code typed for each language.
    document.querySelectorAll('[data-lang]').forEach((button) => {
        button.addEventListener('click', () => {
            const next = button.dataset.lang;
            if (next === lang) return;
            languagePlaceholder[lang] = editor.getValue();
            lang = next;
            document.querySelectorAll('[data-lang]').forEach((item) => {
                item.setAttribute('aria-checked', String(item.dataset.lang === lang));
            });
            editor.setOption('mode', editorMode(lang));
            if (languagePlaceholder[lang] === undefined) {
                languagePlaceholder[lang] = editor.getValue();
            }
            editor.setValue(languagePlaceholder[lang]);
        });
    });

    // Base64 decoding that accepts UTF-8 and the URL-safe alphabet.
    function decodeBase64(value) {
        try {
            let normalized = value.replace(/\s/g, '').replace(/-/g, '+').replace(/_/g, '/');
            normalized += '='.repeat((4 - (normalized.length % 4)) % 4);
            const bytes = Uint8Array.from(atob(normalized), (char) => char.charCodeAt(0));
            return new TextDecoder().decode(bytes);
        } catch {
            return '';
        }
    }

    // Code sent by the embedding page: {type: 'elaina', code: base64, language: {lang: code}}.
    window.addEventListener('message', (event) => {
        if (event.data?.type !== 'elaina') return;
        languagePlaceholder = event.data.language ?? {};
        const code = decodeBase64(event.data.code ?? '');
        if (code !== '') {
            editor.setValue(code);
        }
    });

    // Inside an iframe, offer to continue in a full window with the current code.
    if (window.self !== window.top) {
        const form = $('new-window');
        form.hidden = false;
        form.addEventListener('submit', () => {
            $('post-code').value = editor.getValue();
        });
    }

    // Draggable split between the editor and the output, remembered per browser.
    const workspace = $('workspace');
    const splitter = $('splitter');
    const SPLIT_KEY = 'elaina:runner-split';
    const clamp = (value) => Math.min(85, Math.max(20, value));
    const setSplit = (percent) => {
        const value = clamp(percent);
        workspace.style.setProperty('--editor-size', `${value}%`);
        splitter.setAttribute('aria-valuenow', String(Math.round(value)));
        editor.refresh();
    };
    const saved = Number(localStorage.getItem(SPLIT_KEY));
    if (saved) setSplit(saved);

    const currentSplit = () => parseFloat(getComputedStyle(workspace).getPropertyValue('--editor-size')) || 60;
    const saveSplit = () => localStorage.setItem(SPLIT_KEY, String(currentSplit()));

    splitter.addEventListener('pointerdown', (event) => {
        event.preventDefault();
        splitter.setPointerCapture(event.pointerId);
        document.body.classList.add('is-resizing');
        const move = (moveEvent) => {
            const rect = workspace.getBoundingClientRect();
            setSplit(((moveEvent.clientY - rect.top) / rect.height) * 100);
        };
        const stop = () => {
            splitter.removeEventListener('pointermove', move);
            document.body.classList.remove('is-resizing');
            saveSplit();
        };
        splitter.addEventListener('pointermove', move);
        splitter.addEventListener('pointerup', stop, {once: true});
        splitter.addEventListener('pointercancel', stop, {once: true});
    });
    splitter.addEventListener('keydown', (event) => {
        if (event.key !== 'ArrowUp' && event.key !== 'ArrowDown') return;
        event.preventDefault();
        setSplit(currentSplit() + (event.key === 'ArrowUp' ? -5 : 5));
        saveSplit();
    });
})();
