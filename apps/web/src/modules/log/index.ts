import React from "react";

export const useLog = (ws: WebSocket) => {
    const [log, setLog] = React.useState<string[]>([]);
    React.useEffect(() => {
        const handler = (event: MessageEvent) => {
            setLog([...log, event.data]);
        }
        ws.addEventListener("message", handler)
        return () => {
            ws.removeEventListener("message", handler);
        };
    }, [ws, log]);
    return log;
}
