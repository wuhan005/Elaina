$('#run').click(() => {
    $('#result_bar').css('color', 'white');
    $('#result_data').text('Loading...');

    $.post('', editor.getValue(), (res) => {
        let result = Base64.decode(res.data.result.body);
        let exitCode = res.data.result.exit_code;
        let startAt = res.data.start_at;
        let endAt = res.data.end_at;
        if (exitCode !== 0) {
            $('#result_bar').css('color', 'red');
        }
        $('#result_data').text(result);
        $('#time_cost').text(((endAt - startAt) / 1000000000) + 's');
    }).fail((err) => {
        $('#result_bar').css('color', 'red');
        $('#result_data').text(err.responseJSON.msg);
    })
})
