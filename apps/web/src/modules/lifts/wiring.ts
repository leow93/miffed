import React from "react";
import {initialState, reducer} from "./state.ts";

export const useLiftState = (socket: WebSocket) => {
    const [state, dispatch] = React.useReducer(reducer, initialState)
    React.useEffect(() => {
        socket.onmessage = (event) => {
            const message = JSON.parse(event.data)
            dispatch(message)
        }
    }, [socket])
    return state
}

type SendMessage = {
    type: "call_lift"
    floor: number
}

export const useSendMessage = (socket: WebSocket) => {
    return React.useCallback((message: SendMessage) => {
        socket.send(JSON.stringify(message))
    }, [socket])
}
