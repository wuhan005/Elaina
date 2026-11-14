// Set by iframe.
let languagePlaceholder = {};

const editorMode = (lang) => lang === 'c' ? 'text/x-csrc' : lang;

// Codemirror editor.
const editor = CodeMirror.fromTextArea(document.getElementById('code'), {
    lineNumbers: true,
    mode: editorMode(lang),
    theme: 'material-palenight'
});

$('#run').click(() => {
    // Reset the text color.
    $('#result_bar').css('color', 'white');
    $('#result_data').text('Loading...');

    $.post(window.location.href + '/execute', {'lang': lang, 'code': editor.getValue()}, (res) => {
        $('#result_data').text('');
        const {error, stderr, stdout} = res.data.result;

        if (error) {
            const build_details = $('<details style="color:red;"></details>');
            const build_summary = $('<summary></summary>').text('Error');
            build_details.append(build_summary);
            build_details.append($('<p></p>').text(error));

            $('#result_bar').css('color', 'red');
            $('#result_data').append(build_details);
            $('#result_data').append($('<p></p>').text(stderr));
        } else {
            $('#result_data').text(stdout);
        }

        $('#time_cost').text(((res.data.end_at - res.data.start_at) / 1000000000) + 's');
    }).fail((err) => {
        $('#result_bar').css('color', 'red');
        $('#result_data').text(err.responseJSON?.msg ?? err.statusText);
    })
})

// Switch language
$('[lang]').click((evt) => {
    // Save the current code into language placeholder.
    languagePlaceholder[lang] = editor.getValue();

    // Set the new language.
    lang = evt.target.lang;
    $('#lang').text(lang);
    editor.setOption('mode', editorMode(lang));

    // Recover the language placeholder.
    if (languagePlaceholder[lang] === undefined) {
        languagePlaceholder[lang] = editor.getValue();
    }
    editor.setValue(languagePlaceholder[lang]);
})

// Receive the code from the outside iframe.
window.addEventListener('message', (evt) => {
    if (evt.data.type === 'elaina') {
        languagePlaceholder = evt.data.language
        let code = Base64.decode(evt.data.code ?? '');
        if (code !== '') {
            editor.setValue(code);
            $('#post-code').val(code)
        }
    }
}, false);

// Show button in iframe.
if (window.self !== window.top) {
    $('#new-window').show()
}
