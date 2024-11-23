// Setted by iframe.
let languagePlaceholder = [];

// Codemirror editor.
var editor = CodeMirror.fromTextArea(document.getElementById('code'), {
    lineNumbers: true,
    mode: ((lang) => {
        if (lang === 'c') {
            return 'text/x-csrc'
        }
        return lang
    })(lang),
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
            $('#result_bar').css('color', 'white');
            $('#result_data').text(stdout);
        }

        let startAt = res.data.start_at;
        let endAt = res.data.end_at;
        $('#time_cost').text(((endAt - startAt) / 1000000000) + 's');
    }).fail((err) => {
        $('#result_bar').css('color', 'red');
        $('#result_data').text(err.responseJSON.msg);
    })
})

// Switch language
$('[lang]').click((evt) => {
    // Save the current code into language placeholder.
    languagePlaceholder[lang] = editor.getValue();

    // Set the new language.
    lang = evt.target.lang;
    $('#lang').text(lang);
    editor.setOption('mode', lang);

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
