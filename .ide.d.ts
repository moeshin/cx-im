declare type ApiResponse = {
    ok: boolean
    msg: string
    data: any
}
declare interface ApiRequestOption extends JQuery.AjaxSettings {
    param: object;
}
declare class CxIm {
    static apiRequest(url: string): JQuery.jqXHR;
    static apiRequest(opts: ApiRequestOption): JQuery.jqXHR;
}