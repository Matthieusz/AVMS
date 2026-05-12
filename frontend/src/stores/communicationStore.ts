import { useSyncExternalStore } from 'react';
import type { RSUBeacon, SecurityPolicy } from '@/lib/api';

export type CommStage = 'idle' | 'beacon' | 'auth' | 'join' | 'secure' | 'waiting';

export interface CommunicationEntry {
  vehicleId: string;
  path: 'north' | 'south' | 'east' | 'west';
  lane: 'left' | 'right';
  stage: CommStage;
  message: string;
  sessionKeyRef: string;
  rsuId: string;
  kemAlgorithm: string;
  signatureAlgorithm: string;
  sessionCipher: string;
  bubbleVisible: boolean;
  updatedAt: number;
}

interface CommunicationSnapshot {
  vehicles: CommunicationEntry[];
  policy: SecurityPolicy | null;
  beacons: RSUBeacon[];
  backendStatus: 'loading' | 'connected' | 'error';
  backendMessage: string;
}

const vehicles = new Map<string, CommunicationEntry>();
const listeners = new Set<() => void>();

let policy: SecurityPolicy | null = null;
let beacons: RSUBeacon[] = [];
let backendStatus: CommunicationSnapshot['backendStatus'] = 'loading';
let backendMessage = 'Synchronizing policy and RSU beacons';

let cachedSnapshot: CommunicationSnapshot = {
  vehicles: [],
  policy: null,
  beacons: [],
  backendStatus: 'loading',
  backendMessage,
};

function rebuildSnapshot() {
  cachedSnapshot = {
    vehicles: Array.from(vehicles.values()).sort((left, right) => right.updatedAt - left.updatedAt),
    policy,
    beacons,
    backendStatus,
    backendMessage,
  };
}

function emitChange() {
  rebuildSnapshot();
  for (const listener of listeners) {
    listener();
  }
}

function getSnapshot(): CommunicationSnapshot {
  return cachedSnapshot;
}

export function subscribeCommunicationStore(listener: () => void) {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function useCommunicationSnapshot() {
  return useSyncExternalStore(subscribeCommunicationStore, getSnapshot, getSnapshot);
}

export function registerVehicleCommunication(entry: Omit<CommunicationEntry, 'updatedAt'>) {
  vehicles.set(entry.vehicleId, { ...entry, updatedAt: Date.now() });
  emitChange();
}

export function updateVehicleCommunication(vehicleId: string, patch: Partial<Omit<CommunicationEntry, 'vehicleId' | 'path' | 'lane'>>) {
  const current = vehicles.get(vehicleId);
  if (!current) {
    return;
  }

  const next: CommunicationEntry = {
    ...current,
    ...patch,
    updatedAt: Date.now(),
  };

  if (
    next.stage === current.stage &&
    next.message === current.message &&
    next.sessionKeyRef === current.sessionKeyRef &&
    next.rsuId === current.rsuId &&
    next.kemAlgorithm === current.kemAlgorithm &&
    next.signatureAlgorithm === current.signatureAlgorithm &&
    next.sessionCipher === current.sessionCipher &&
    next.bubbleVisible === current.bubbleVisible
  ) {
    return;
  }

  vehicles.set(vehicleId, next);
  emitChange();
}

export function unregisterVehicleCommunication(vehicleId: string) {
  if (vehicles.delete(vehicleId)) {
    emitChange();
  }
}

export function setCommunicationSecurityProfile(nextPolicy: SecurityPolicy, nextBeacons: RSUBeacon[]) {
  policy = nextPolicy;
  beacons = nextBeacons;
  backendStatus = 'connected';
  backendMessage = 'Policy and RSU keys synchronized with backend';
  emitChange();
}

export function setCommunicationBackendError(message: string) {
  backendStatus = 'error';
  backendMessage = message;
  emitChange();
}

export function getVehicleCommunication(vehicleId: string) {
  return vehicles.get(vehicleId) ?? null;
}