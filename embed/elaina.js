function Elaina(obj) {
    let host = obj.host.replace(/\/+$/, '');
    let uid = obj.uid;
    let el = obj.el;
    let language = obj.language;
    let height = obj.height;

    let frame = document.createElement('iframe');
    frame.src = host + '/r/' + uid;
    frame.width = '100%';
    frame.height = height;
    frame.frameBorder = 0;
    frame.onload = () => {
        frame.contentWindow.postMessage({
            type: 'elaina',
            language: language,
        }, host)
    }

    let elainaElement = document.getElementById(el);
    elainaElement.appendChild(frame);
}