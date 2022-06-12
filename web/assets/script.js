const MDUI_DIALOG_OPTIONS = {
    modal: true,
    history: false,
};
function emptyFunc() {}
function isSwitchClickOnList(e) {
    return e.target.classList.contains('mdui-switch')
        || e.target.parentElement.classList.contains('mdui-switch');
}
/**
 * @param {JQuery} $list
 * @param {function(JQuery)} onClick
 * @param {function(JQuery)} onLongClick
 */
function onListClick($list, onClick, onLongClick) {
    const selector = '.mdui-list-item';
    let timer = -1;
    let target;
    let isMove = false;
    let x = 0;
    let y = 0;
    function stopTimer() {
        if (timer !== -1) {
            clearTimeout(timer);
            timer = -1;
        }
    }
    function p(e, p) {
        return e['offset' + p]
            || (e.changedTouches && e.changedTouches.length && e.changedTouches[0]['client' + p]) || 0;
    }
    function jq(e) {
        let $this = $(e.currentTarget);
        if ($this.is(selector)) {
            return $this;
        }
        return $this.parents(selector).last();
    }
    $list.on('mousemove', e => {
        if ((x !== p(e, 'X') || y !== p(e, 'Y')) && !isMove) {
            stopTimer();
            isMove = true;
        }
    }).on('mousedown', selector, e => {
        stopTimer();
        // noinspection JSDeprecatedSymbols
        if (e.which !== 1) {
            return;
        }
        x = p(e, 'X');
        y = p(e, 'Y');
        if (isSwitchClickOnList(e)) {
            return;
        }
        target = e.currentTarget;
        isMove = false;
        timer = setTimeout(() => {
            target = null;
            stopTimer();
            if (isMove) {
                return;
            }
            onLongClick(jq(e));
        }, 300);
    }).on('mouseup', selector, e => {
        stopTimer();
        // noinspection JSDeprecatedSymbols
        if (!onClick || isMove || e.which !== 1 || target !== e.currentTarget || isSwitchClickOnList(e)) {
            return;
        }
        return onClick(jq(e));
    }).on('contextmenu', selector, e => {
        e.preventDefault();
        stopTimer();
        if (!onLongClick) {
            return;
        }
        return onLongClick(jq(e));
    });
}
class CxIm {
    static apiRequest(opts) {
        if (typeof opts !== 'object') {
            opts = {url: opts};
        } else {
            const param = opts.param;
            if (param) {
                const query = $.param(param);
                if (query) {
                    let url = opts.url;
                    const end = url.length - 1;
                    const i = url.lastIndexOf('?');
                    if (i === -1) {
                        url += '?';
                    } else if (i !== end) {
                        const i = url.lastIndexOf('&');
                        if (i === -1 || i !== end) {
                            url += '&';
                        }
                    }
                    url += query;
                    opts.url = url;
                }
            }
        }

        opts.dataType = 'json';
        return $.ajax(opts).then(this.handleApiResponse);
    }
    /** @param {ApiResponse} data */
    static handleApiResponse(data) {
        const msg = data.msg;
        if (data.ok) {
            if (msg) {
                console.warn(msg);
            }
            return data.data;
        }
        mainAlert(msg, '请求失败');
        return Promise.reject(msg);
    }
}

function getDialogOptions(options, key, val) {
    let o;
    if (options) {
        o = $.extend({}, MDUI_DIALOG_OPTIONS, options);
    }
    if (val) {
        if (!o) {
            o = $.extend({}, MDUI_DIALOG_OPTIONS);
        }
        o[key] = val;
    }
    if (!o) {
        o = MDUI_DIALOG_OPTIONS;
    }
    return o;
}
function mainAlert(text, title, btn, callback = emptyFunc, options) {
    return mdui.alert(text, title, callback, getDialogOptions(options, 'confirmText', btn));
}
