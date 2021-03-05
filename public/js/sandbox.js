// Setted by iframe.
let languagePlaceholder = [];

// Codemirror editor.
var editor = CodeMirror.fromTextArea(document.getElementById('code'), {
    lineNumbers: true,
    mode: lang,
    theme: 'material-palenight'
});

$('#run').click(() => {
    // Reset the text color.
    $('#result_bar').css('color', 'white');

    $('#result_data').text('Loading...');

    $.post('', {'lang': lang, 'code': editor.getValue()}, (res) => {
        $('#result_data').text('');
        let steps = res.data.result;

        steps.forEach((item, index) => {
            item.body = Base64.decode(item.body)
            if(!item.error && index === 0){
                let build_details = $('<details style="color:gray;"></details>');
                let build_summary = $('<summary></summary>').text('Build logs');
                build_details.append(build_summary);
                build_details.append($('<p></p>').text(item.body));
                $('#result_data').append(build_details);
            }else {
                $('#result_data').append(item.body)
                if (item.error) {
                    $('#result_bar').css('color', 'red');
                }
            }
        })

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
        }
    }
}, false);