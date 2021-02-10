$('#run').click(() => {
    way.set('result', 'Loading...');
    $.post('', editor.getValue(), (res) => {
        way.set('result', atob(res.data.result.body));
    })
})
