$('#run').click(() => {
    $('#result_bar').css('color', 'white');
    $('#result_data').text('Loading...');

    $.post('', editor.getValue(), (res) => {
        let result = atob(res.data.result.body);
        let exitCode = res.data.result.exit_code;
        if (exitCode !== 0) {
            $('#result_bar').css('color', 'red');
        }
        $('#result_data').text(result);
    }).fail((err) => {
        $('#result_bar').css('color', 'red');
        $('#result_data').text(err.responseJSON.msg);
    })
})
