import { it, describe, expect } from "vitest";
import { createEntityState, pushEntity, updateEntityBy } from "./entity-state.ts";

type testType = {
  id: string;
  foo: string;
};

describe("create entity state", () => {
  it("creates an empty state when an empty array is passed", () => {
    expect(createEntityState<testType, string>([])).toEqual({
      ids: [],
      entities: {},
    });
  });

  it("correctly creates state", () => {
    expect(
      createEntityState<testType, string>([
        { id: "1", foo: "bar" },
        { id: "2", foo: "baz" },
      ]),
    ).toEqual({
      ids: ["1", "2"],
      entities: {
        1: {
          id: "1",
          foo: "bar",
        },
        2: { id: "2", foo: "baz" },
      },
    });
  });
});

describe("pushEntity", () => {
  it("appends an object to the end of the list", () => {
    const state = createEntityState<testType, string>([
      { id: "1", foo: "bar" },
      { id: "2", foo: "baz" },
    ]);

    expect(pushEntity(state, { id: "3", foo: "boo" })).toEqual({
      ids: ["1", "2", "3"],
      entities: {
        1: {
          id: "1",
          foo: "bar",
        },
        2: { id: "2", foo: "baz" },
        3: { id: "3", foo: "boo" },
      },
    });
  });
});

describe("updateEntityBy", () => {
  it("does nothing if the entity is not found", () => {
    const state = createEntityState<testType, string>([
      { id: "1", foo: "bar" },
      { id: "2", foo: "baz" },
    ]);

    expect(updateEntityBy(state, "3", x => ({ id: x.id, foo: "UPDATED" }))).toEqual(state);
  });

  it("updates according to the given function", () => {
    const state = createEntityState<testType, string>([
      { id: "1", foo: "bar" },
      { id: "2", foo: "baz" },
    ]);

    expect(updateEntityBy(state, "2", x => ({ id: x.id, foo: "UPDATED" }))).toEqual({
      ids: ["1", "2"],
      entities: {
        1: {
          id: "1",
          foo: "bar",
        },
        2: { id: "2", foo: "UPDATED" },
      },
    });
  });

  it("entity id cannot be overwritten", () => {
    const state = createEntityState<testType, string>([
      { id: "1", foo: "bar" },
      { id: "2", foo: "baz" },
    ]);

    expect(updateEntityBy(state, "2", x => ({ ...x, id: "heheh" }))).toEqual({
      ids: ["1", "2"],
      entities: {
        1: {
          id: "1",
          foo: "bar",
        },
        2: { id: "2", foo: "baz" },
      },
    });
  });
});
