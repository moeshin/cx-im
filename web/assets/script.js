$ = mdui.$;
class CxIm {
    static apiRequest(opts) {
        if (typeof opts !== 'object') {
            opts = {url: opts};
        }
        opts.dataType = 'json';
        return $.ajax(opts)
            .then(data => {
                const msg = data['msg'];
                if (data.ok) {
                    if (msg) {
                        console.warn(msg);
                    }
                    return data.data;
                }
                return Promise.reject(msg);
            });
    }
}