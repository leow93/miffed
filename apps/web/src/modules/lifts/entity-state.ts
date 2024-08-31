type Id = string | number;

export type EntityState<K extends Id, T extends { id: K }> = {
  ids: K[];
  entities: Record<K, T>;
};

export const createEntityState = <T extends { id: K }, K extends Id = string>(
  entities: T[],
): EntityState<K, T> => {
  const ids: K[] = Array.from({ length: entities.length });
  const entitiesRecord = {} as Record<K, T>;
  for (let i = 0; i < entities.length; i++) {
    const entity = entities[i];
    ids[i] = entity.id;
    entitiesRecord[entity.id] = entity;
  }

  return {
    ids,
    entities: entitiesRecord,
  };
};

export const pushEntity = <T extends { id: K }, K extends Id = string>(
  state: EntityState<K, T>,
  item: T,
): EntityState<K, T> => {
  const entities = {
    ...state.entities,
    [item.id]: item,
  };
  return {
    ids: state.ids.concat(item.id),
    entities,
  };
};

export const updateEntityBy = <T extends { id: K }, K extends Id = string>(
  state: EntityState<K, T>,
  id: K,
  fn: (x: T) => T,
): EntityState<K, T> => {
  const entity = state.entities[id];
  if (!entity) return state;

  return {
    ids: state.ids,
    entities: {
      ...state.entities,
      [id]: {
        ...fn(entity),
        id: id,
      },
    },
  };
};
