const API_URL = "http://localhost:8080";
type LiftId = string;
type Lift = {
  id: string;
  floor: number;
};

export const getLifts = (): Promise<Lift[]> => fetch(API_URL + "/lift").then(res => res.json());
export const createLift = (opts: { floor: number }): Promise<Lift> =>
  fetch(API_URL + "/lift", {
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(opts),
    method: "POST",
  }).then(res => res.json());

export const callLift = (id: LiftId, floor: number) =>
  fetch(API_URL + `/lift/${id}/call`, {
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      floor,
    }),
    method: "POST",
  });
