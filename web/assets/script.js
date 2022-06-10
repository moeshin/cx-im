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