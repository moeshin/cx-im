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
            const obj = opts.params;
            if (typeof obj === 'object') {
                let url = new URL(opts.url, location.href);
                const params = url.searchParams;
                for (const k in obj) {
                    params.set(k, obj[k]);
                }
                opts.url = url.toString();
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
        return Promise.reject(msg);
    }
}