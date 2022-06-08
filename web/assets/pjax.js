/*
 * Pjax.js 0.6.4
 *
 * Copyright (c) 2020 Guilherme Nascimento (brcontainer@yahoo.com.br)
 *
 * Released under the MIT license
 */

(function (u) {
    "use strict";

    let xhr, config, timer, loader,
        w = typeof window !== "undefined" ? window : {},
        d = w.document || {},
        history = w.history,
        URL = w.URL,
        domparser = !!w.DOMParser,
        formdata = !w.FormData,
        evts = {},
        docImplementation = d.implementation,
        location = w.location,
        host = location ? (location.protocol.replace(/:/g, "") + "://" + w.location.host) : '',
        inputRe = /^(input|textarea|select|datalist|button|output)$/i,
        started = false,
        elementProto = w.Element && w.Element.prototype,
        ArraySlice = Array.prototype.slice,
        PUSH = 1,
        REPLACE = 2,
        supported = (
            elementProto && history.pushState && (
                domparser || (docImplementation && docImplementation.createHTMLDocument)
            )
        );

    const main = {
        "supported": supported,
        "remove": remove,
        "start": start,
        "request": function (url, cfg) {
            pjaxLoad(url, cfg.replace ? REPLACE : PUSH, cfg.method, null, cfg.data);
        },
        "on": function (name, callback) {
            pjaxEvent(name, callback);
        },
        "off": function (name, callback) {
            pjaxEvent(name, callback, true);
        },
        "loadUrl": function (url) {
            pjaxLoad(url, PUSH, 'GET', null, null, config);
        }
    };

    if (domparser) {
        try {
            const test = parseHtml('<html lang="en"><head><title>1</title></head><body>1</body></html>');
            domparser = !!(test.head && test.title && test.body);
        } catch (ee) {
            domparser = false;
        }
    }

    function showLoader()
    {
        if (timer) {
            clearTimeout(timer);
            timer = 0;
            loader.className = "pjax-loader pjax-hide";
            setTimeout(showLoader, 20);
            return;
        }

        if (!loader) {
            loader = d.createElement("div");
            loader.innerHTML = '<div class="pjax-progress"></div>';
            d.body.appendChild(loader);
        }

        loader.className = "pjax-loader pjax-start";

        timer = setTimeout(function () {
            timer = 0;
            loader.className += " pjax-inload";
        }, 10);
    }

    function hideLoader()
    {
        if (timer) clearTimeout(timer);

        loader.className += " pjax-end";

        timer = setTimeout(function () {
            timer = 0;
            loader.className += " pjax-hide";
        }, 1000);
    }

    function selectorEach(context, query, callback)
    {
        [].slice.call(context.querySelectorAll(query)).forEach(callback);
    }

    function serialize(form)
    {
        const data = [];

        selectorEach(form, "[name]", function (el) {
            if (el.name && inputRe.test(el.tagName)) {
                data.push(encodeURIComponent(el.name) + "=" + encodeURIComponent(el.value));
            }
        });

        return data.join("&");
    }

    function parseHtml(data)
    {
        if (domparser) return (new DOMParser).parseFromString(data, "text/html");

        let pd = docImplementation.createHTMLDocument("");

        if (/^\s*<(!doctype|html)[^>]*>/i.test(data)) {
            pd.documentElement.innerHTML = data;

            if (!pd.body || !pd.head) {
                pd = docImplementation.createHTMLDocument("");
                pd.write(data);
            }
        } else {
            pd.body.innerHTML = data;
        }

        return pd;
    }

    function pjaxUpdateHead(head)
    {
        let i, index, frag, nodes = [], j = head.children;

        for (i = j.length - 1; i >= 0; i--) {
            if (j[i].tagName !== "TITLE") nodes.push(j[i].outerHTML);
        }

        j = d.head.children;

        for (i = j.length - 1; i >= 0; i--) {
            if (j[i].tagName === "TITLE" || getData(j[i], "resource")) continue;

            index = nodes.indexOf(j[i].outerHTML);

            if (index === -1) {
                j[i].parentNode.removeChild(j[i]);
            } else {
                nodes.splice(index, 1);
            }
        }

        frag = d.createElement("div");
        frag.innerHTML = nodes.join("");

        j = frag.children;

        for (i = j.length - 1; i >= 0; i--) {
            if (!getData(j[i], "resource")) d.head.appendChild(j[i]);
        }

        frag = nodes = null;
    }

    function pjaxTrigger(name, arg1, arg2, arg3)
    {
        if (evts[name]) {
            let ce = evts[name], i = 0, j = ce.length;
            for (; i < j; i++) ce[i](arg1, arg2, arg3);
        }
    }

    function pjaxEvent(name, callback, remove)
    {
        if (typeof callback === "function") {
            if (!evts[name]) evts[name] = [];

            if (remove) {
                evts[name] = evts[name].filter(function (item) {
                    return item !== callback;
                });
            } else {
                evts[name].push(callback);
            }
        }
    }

    function pjaxParse(url, data, cfg, state)
    {
        pjaxTrigger('replace', url);

        let current, tmp = parseHtml(data);

        let title = tmp.title || "",
            containers = cfg.containers,
            insertion = cfg.insertion,
            x = cfg.scrollLeft > 0 ? +cfg.scrollLeft : w.scrollX || w.pageXOffset,
            y = cfg.scrollTop > 0 ? +cfg.scrollTop : w.scrollY || w.pageYOffset;

        if (state) {
            const c = {
                "pjaxUrl": url,
                "pjaxData": data,
                "pjaxConfig": cfg
            };

            if (state === PUSH) {
                history.pushState(c, title, url);
            } else if (state === REPLACE) {
                history.replaceState(c, title, url);
            }
        }

        if (evts.dom) return pjaxTrigger("dom", url, tmp);

        if (cfg.updatehead && tmp.head) pjaxUpdateHead(tmp.head);

        d.title = title;

        for (let i = containers.length - 1; i >= 0; i--) {
            current = tmp.body.querySelector(containers[i]);

            if (current) {
                selectorEach(d, containers[i], function (el) {
                    if (insertion === "append" || insertion === "prepend") {
                        let fragment = d.createDocumentFragment();

                        let i = 0, nodes = ArraySlice.call(current.childNodes), j = nodes.length;
                        for (; i < j; ++i) {
                            fragment.appendChild(nodes[i]);
                        }

                        if (insertion === "append") {
                            el.appendChild(fragment);
                        } else {
                            el.insertBefore(fragment, el.firstChild);
                        }

                        fragment = null;
                    } else {
                        el.innerHTML = current.innerHTML;
                    }
                    const els = el.querySelectorAll('script');
                    for (let el of els) {
                        evalScript(el);
                    }
                });
            }
        }

        w.scrollTo(x, y);

        tmp = containers = null;
    }

    function getData(el, name)
    {
        let data = el.getAttribute("data-pjax-" + name);

        if (data === "true" || data === "false") {
            return data === "true";
        } else if (!isNaN(data)) {
            return parseFloat(data);
        } else if (/^\[[\s\S]+]$|^{[^:]+[:][\s\S]+}$/.test(data)) {
            try {
                data = JSON.parse(data);
            } catch (e) {}
        }

        return data;
    }

    function pjaxAttributes(el)
    {
        let current, value, cfg = JSON.parse(JSON.stringify(config)),
            attrs = [
                "containers", "updatecurrent", "updatehead", "insertion",
                "loader", "scroll-left", "scroll-top", "done", "fail"
            ];

        if (!el) return cfg;

        for (let i = attrs.length - 1; i >= 0; i--) {
            current = attrs[i];

            value = getData(el, current);

            if (value) {
                current = current.toLowerCase().replace(/-([a-z])/g, function (a, b) {
                    return b.toUpperCase();
                });

                cfg[current] = value;
            }
        }

        return cfg;
    }

    function pjaxFinish(url, cfg, state, el, callback, data)
    {
        if (cfg.loader) hideLoader();

        if (data) pjaxParse(url, data, cfg, state);

        pjaxTrigger(data ? "done" : "fail", url);
        pjaxTrigger("then", url);

        if (callback) new Function(callback).call(el);
    }

    function pjaxAbort()
    {
        if (xhr) xhr.abort();
    }

    function pjaxNoCache(url)
    {
        let u, n = "_=" + (+new Date);

        if (!URL) {
            u = new URL(url);
        } else {
            u = d.createElement("a");
            u.href = url;
        }

        u.search += (u.search ? "&" : "?") + n;

        url = u + "";
        u = null;

        return url;
    }

    function pjaxLoad(url, state, method, el, data, cfg)
    {
        pjaxAbort();

        pjaxTrigger("initiate", url, cfg);

        if (cfg.loader) showLoader();

        const headers = config.headers;

        headers["X-PJAX-Container"] = cfg.containers.join(",");
        headers["X-PJAX"] = "true";

        if (evts.handler) {
            return pjaxTrigger("handler", {
                "url": url,
                "state": state,
                "method": method,
                "element": el
            }, cfg, pjaxFinish);
        }

        if (config.proxy) url = config.proxy + encodeURIComponent(url);

        if (config.nocache) url = pjaxNoCache(url);

        xhr = new XMLHttpRequest;
        xhr.open(method, url, true);

        for (let k in headers) xhr.setRequestHeader(k, headers[k]);

        xhr.onreadystatechange = function () {
            if (this.readyState !== 4) return;

            const status = this.status;

            if (status >= 200 && status < 300) {
                const containers = xhr.getResponseHeader("X-PJAX-Container");

                if (containers) cfg.containers = containers.split(",");

                pjaxFinish(xhr.getResponseHeader("X-PJAX-URL") || url, cfg, state, el, cfg.done, this.responseText);
            } else {
                pjaxFinish(url, cfg, status, el, cfg.fail);
            }
        };

        xhr.send(data || "");
    }

    function pjaxLink(e)
    {
        if (e.button === 0) {
            let url, lastEl, el = e.target;

            if (el.matches(config.linkSelector)) {
                url = el.href;
            } else {
                while ((el = el.parentNode)) {
                    if (el.tagName === "A") {
                        lastEl = el;
                        break;
                    }
                }

                if (!lastEl || !lastEl.matches(config.linkSelector)) return;

                el = lastEl;
                url = el.href;
            }

            pjaxRequest("GET", url, null, el, e);
        }
    }

    function pjaxForm(e)
    {
        let url, data, method, el = e.target;

        if (!el.matches(config.formSelector)) return;

        method = String(el.method).toUpperCase();

        if (method === "POST" && !formdata) return;

        url = el.action;

        if (method !== "POST" || el.enctype !== "multipart/form-data") {
            data = serialize(el);

            if (method !== "POST") {
                url = url.replace(/\?.*/g, "") + "?";
                if (data) url += data;
                data = null;
            }
        } else if (formdata) {
            data = new FormData(el);
        } else {
            return;
        }

        pjaxRequest(method, url, data, el, e);
    }

    function pjaxRequest(method, url, data, el, e)
    {
        if (url === host || url.indexOf(host + "/") === 0) {
            const target = String(el.target).toLowerCase();

            if (!target || target === "_self" || target === w.name || (target === "_parent" && w === w.parent)) {
                e.preventDefault();

                const cfg = pjaxAttributes(el);

                if (method === "POST" || url !== w.location + "") {
                    pjaxLoad(url, PUSH, method, el, data, cfg);
                } else if (cfg.updatecurrent) {
                    pjaxLoad(url, REPLACE, method, el, data, cfg);
                }
            }
        }
    }

    function pjaxState(e)
    {
        if (e.state && e.state.pjaxUrl) {
            pjaxAbort();
            pjaxParse(e.state.pjaxUrl, e.state.pjaxData, e.state.pjaxConfig, false);
            pjaxTrigger("history", e.state.pjaxUrl, e.state);
        }
    }

    function ready()
    {
        if (!started) {
            started = true;

            const url = w.location + "", state = w.history.state;

            if (!state || !state.pjaxUrl) {
                history.replaceState({
                    "pjaxUrl": url,
                    "pjaxData": d.documentElement.outerHTML,
                    "pjaxConfig": config
                }, d.title, url);
            }

            w.addEventListener("unload", pjaxAbort);
            w.addEventListener("popstate", pjaxState);

            if (config.linkSelector) d.addEventListener("click", pjaxLink);
            if (config.formSelector) d.addEventListener("submit", pjaxForm);
        }
    }

    function remove()
    {
        if (config && started) {
            d.removeEventListener("click", pjaxLink);
            d.removeEventListener("submit", pjaxForm);

            w.removeEventListener("unload", pjaxAbort);
            w.removeEventListener("popstate", pjaxState);

            started = false;
            config = u;
        }
    }

    function start(opts)
    {
        if (supported && /^https?:$/.test(w.location.protocol)) {
            remove();

            config = {
                "linkSelector": "a:not([data-pjax-ignore]):not([href^='#']):not([href^='javascript:'])",
                "formSelector": "form:not([data-pjax-ignore]):not([action^='javascript:'])",
                "containers": [ "#pjax-container" ],
                "updatecurrent": false,
                "updatehead": true,
                "insertion": true,
                "scrollLeft": 0,
                "scrollTop": 0,
                "nocache": false,
                "loader": true,
                "proxy": "",
                "headers": {}
            };

            for (let k in config) {
                if (opts && k in opts) config[k] = opts[k];
            }

            opts = null;

            if (/^(interactive|complete)$/.test(d.readyState)) {
                ready();
            } else {
                d.addEventListener("DOMContentLoaded", ready);
            }
        }
    }

    function evalScript(el) {
        const code = el.text || el.textContent || el.innerHTML || "";
        const src = el.src || "";
        const parent =
            el.parentNode || document.querySelector("head") || document.documentElement;
        const script = document.createElement("script");

        if (code.match("document.write")) {
            if (console && console.log) {
                console.log(
                    "Script contains document.write. Can’t be executed correctly. Code skipped ",
                    el
                );
            }
            return false;
        }

        script.type = "text/javascript";
        script.id = el.id;

        /* istanbul ignore if */
        if (src !== "") {
            script.src = src;
            script.async = false; // force synchronous loading of peripheral JS
        }

        if (code !== "") {
            try {
                script.appendChild(document.createTextNode(code));
            } catch (e) {
                /* istanbul ignore next */
                // old IEs have funky script nodes
                script.text = code;
            }
        }

        // execute
        parent.appendChild(script);
        // avoid pollution only in head or body tags
        if (
            (parent instanceof HTMLHeadElement || parent instanceof HTMLBodyElement) &&
            parent.contains(script)
        ) {
            parent.removeChild(script);
        }

        return true;
    }

    if (elementProto && !elementProto.matches) {
        // noinspection JSUnresolvedVariable
        elementProto.matches = elementProto.matchesSelector || elementProto.mozMatchesSelector
            || elementProto.msMatchesSelector || elementProto.oMatchesSelector || elementProto.webkitMatchesSelector
            || function (query) {
                let els = (this.document || this.ownerDocument).querySelectorAll(query), i = els.length;

                while (--i >= 0 && els[i] !== this) {}
                return i > -1;
            };
    }

    w.Pjax = main;

    // CommonJS
    // if (typeof module !== "undefined" && module.exports) module.exports = main;

    // RequireJS
    // if (typeof define !== "undefined") define(function () { return main; });
})();
