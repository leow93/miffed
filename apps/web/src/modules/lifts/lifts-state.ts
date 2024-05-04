import {LiftState, Message, reducer as liftReducer} from './lift-state.ts'

type State = Record<string, LiftState>

type Action =
    { type: 'initialise'; data: Record<string, { currentFloor: number; lowestFloor: number; highestFloor: number }> }
    | Message & { liftId: string }

export const initialState: State = {}

export const reducer = (state: State, action: Action): State => {
    if (action.type === 'initialise') {
        return Object.keys(action.data).reduce((acc, liftId) => {
            const data = action.data[liftId]
            acc[liftId] = {
                type: 'created',
                currentFloor: data.currentFloor,
                lowestFloor: data.lowestFloor,
                highestFloor: data.highestFloor,
            }
            return acc
        }, {} as State)
    }


    const liftState = state[action.liftId]
    if (!liftState) {
        return state
    }
    const newLiftState = liftReducer(liftState, action)
    return {...state, [action.liftId]: newLiftState}
}