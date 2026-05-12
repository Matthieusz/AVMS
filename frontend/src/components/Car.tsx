import { useRef, useMemo, useEffect } from 'react';
import { useFrame } from '@react-three/fiber';
import { Box, Html } from '@react-three/drei';
import * as THREE from 'three';
import { getLightColorForPath } from '@/stores/trafficStore';
import { registerCar, updateCar, unregisterCar, getCarAhead } from '@/stores/carStore';
import {
  registerVehicleCommunication,
  unregisterVehicleCommunication,
  useCommunicationSnapshot,
  updateVehicleCommunication,
} from '@/stores/communicationStore';

interface CarProps {
  vehicleId: string;
  path: 'north' | 'south' | 'east' | 'west';
  lane: 'left' | 'right';
  speed?: number;
  color?: string;
  startOffset?: number;
}

const LANE_OFFSET = 1.5;
const ROAD_RANGE = 60;
const STOP_LINE_PROGRESS = 25;
const SAFE_FOLLOW_DISTANCE = 12;
const FOLLOW_BUFFER = 3;
const BRAKE_ZONE = 14;

export function Car({ vehicleId, path, lane, speed: baseSpeed = 4, color = '#ef4444', startOffset = 0 }: CarProps) {
  const ref = useRef<THREE.Group>(null);
  const progress = useRef(startOffset);
  const currentSpeed = useRef(baseSpeed);
  const carId = useRef(vehicleId);
  const communicationSnapshot = useCommunicationSnapshot();

  const { startPos, direction, axis } = useMemo(() => {
    const offset = lane === 'left' ? -LANE_OFFSET : LANE_OFFSET;
    switch (path) {
      case 'north':
        return { startPos: new THREE.Vector3(offset, 0.4, 30), direction: -1, axis: 'z' as const };
      case 'south':
        return { startPos: new THREE.Vector3(-offset, 0.4, -30), direction: 1, axis: 'z' as const };
      case 'east':
        return { startPos: new THREE.Vector3(-30, 0.4, offset), direction: 1, axis: 'x' as const };
      case 'west':
        return { startPos: new THREE.Vector3(30, 0.4, -offset), direction: -1, axis: 'x' as const };
    }
  }, [path, lane]);

  useEffect(() => {
    const id = carId.current;
    registerCar(id, path, lane);
    registerVehicleCommunication({
      vehicleId,
      path,
      lane,
      stage: 'idle',
      message: 'Scanning for RSU beacons',
      sessionKeyRef: createSessionKeyRef(vehicleId),
      rsuId: rsuIdForPath(path),
      kemAlgorithm: 'ML-KEM-768',
      signatureAlgorithm: 'ML-DSA',
      sessionCipher: 'AES-256-GCM',
      bubbleVisible: false,
    });
    return () => {
      unregisterCar(id);
      unregisterVehicleCommunication(vehicleId);
    };
  }, [lane, path, vehicleId]);

  useFrame((_, delta) => {
    if (!ref.current) return;

    const lightColor = getLightColorForPath(path);
    const carAhead = getCarAhead(carId.current, path, lane);

    const wrappedProgress = ((progress.current % ROAD_RANGE) + ROAD_RANGE) % ROAD_RANGE;

    let targetSpeed = baseSpeed;

    // Traffic light stop logic
    const distToStopLine = (STOP_LINE_PROGRESS - wrappedProgress + ROAD_RANGE) % ROAD_RANGE;
    const isBeforeStopLine = distToStopLine > 0 && distToStopLine < BRAKE_ZONE + 2;
    const hasPassedStopLine = wrappedProgress >= STOP_LINE_PROGRESS;

    if (!hasPassedStopLine && isBeforeStopLine) {
      if (lightColor === 'red') {
        const factor = Math.max(0, (distToStopLine - 1) / BRAKE_ZONE);
        targetSpeed = baseSpeed * factor;
      } else if (lightColor === 'yellow') {
        // Continue through if very close to stop line (can't stop safely)
        if (distToStopLine > 2.5) {
          const factor = Math.max(0, (distToStopLine - 1) / BRAKE_ZONE);
          targetSpeed = baseSpeed * factor;
        }
      }
    }

    // Car ahead collision avoidance
    if (carAhead) {
      if (carAhead.distance < SAFE_FOLLOW_DISTANCE) {
        const factor = Math.max(
          0,
          (carAhead.distance - FOLLOW_BUFFER) / (SAFE_FOLLOW_DISTANCE - FOLLOW_BUFFER),
        );
        targetSpeed = Math.min(targetSpeed, baseSpeed * factor);
      }
    }

    // Smooth speed transition
    const lerpFactor = Math.min(delta * 4, 1);
    currentSpeed.current += (targetSpeed - currentSpeed.current) * lerpFactor;

    // Update progress
    progress.current += currentSpeed.current * delta;

    // Update store
    updateCar(carId.current, progress.current);

    // Update position
    const wrapped = ((progress.current % ROAD_RANGE) + ROAD_RANGE) % ROAD_RANGE;
    const pos = -ROAD_RANGE / 2 + wrapped;

    if (axis === 'z') {
      ref.current.position.set(startPos.x, startPos.y, pos * direction);
      ref.current.rotation.set(0, direction > 0 ? Math.PI : 0, 0);
    } else {
      ref.current.position.set(pos * direction, startPos.y, startPos.z);
      ref.current.rotation.set(0, direction > 0 ? -Math.PI / 2 : Math.PI / 2, 0);
    }

    const communicationState = deriveCommunicationState({
      vehicleId,
      path,
      wrappedProgress: wrapped,
      lightColor,
      moving: currentSpeed.current > 0.45,
    });

    updateVehicleCommunication(vehicleId, communicationState);
  });

  const communication = communicationSnapshot.vehicles.find((entry) => entry.vehicleId === vehicleId) ?? null;

  return (
    <group ref={ref} position={startPos.toArray()} castShadow>
      {communication?.bubbleVisible ? (
        <Html position={[0, 1.75, 0]} transform sprite distanceFactor={22} pointerEvents="none" zIndexRange={[400, 0]}>
          <div className="max-w-[7.5rem] rounded-md border border-cyan-300/35 bg-slate-950/85 px-2 py-1 text-center shadow-[0_8px_18px_rgba(2,6,23,0.38)] backdrop-blur-sm">
            <div className="truncate text-[8px] font-semibold uppercase tracking-[0.18em] text-cyan-300">
              {communication.vehicleId}
            </div>
            <div className="truncate text-[9px] leading-tight text-white">{communication.message}</div>
          </div>
        </Html>
      ) : null}

      {/* Body */}
      <Box args={[1.6, 0.7, 3.2]} position={[0, 0.35, 0]} castShadow>
        <meshStandardMaterial color={color} roughness={0.3} metalness={0.2} />
      </Box>
      {/* Cabin */}
      <Box args={[1.3, 0.5, 1.8]} position={[0, 0.9, -0.1]} castShadow>
        <meshStandardMaterial color="#1f2937" roughness={0.1} metalness={0.5} />
      </Box>
      {/* Headlights */}
      <Box args={[0.3, 0.15, 0.05]} position={[-0.5, 0.4, 1.61]}>
        <meshStandardMaterial color="#fef08a" emissive="#fef08a" emissiveIntensity={2} />
      </Box>
      <Box args={[0.3, 0.15, 0.05]} position={[0.5, 0.4, 1.61]}>
        <meshStandardMaterial color="#fef08a" emissive="#fef08a" emissiveIntensity={2} />
      </Box>
      {/* Taillights */}
      <Box args={[0.3, 0.15, 0.05]} position={[-0.5, 0.45, -1.61]}>
        <meshStandardMaterial color="#dc2626" emissive="#dc2626" emissiveIntensity={1.5} />
      </Box>
      <Box args={[0.3, 0.15, 0.05]} position={[0.5, 0.45, -1.61]}>
        <meshStandardMaterial color="#dc2626" emissive="#dc2626" emissiveIntensity={1.5} />
      </Box>
      {/* Wheels */}
      {[[-0.85, 0.2, 1], [0.85, 0.2, 1], [-0.85, 0.2, -1], [0.85, 0.2, -1]].map((p, i) => (
        <Box key={i} args={[0.25, 0.4, 0.4]} position={p as [number, number, number]} castShadow>
          <meshStandardMaterial color="#111827" roughness={0.9} />
        </Box>
      ))}
    </group>
  );
}

function rsuIdForPath(path: CarProps['path']) {
  switch (path) {
    case 'north':
      return 'rsu-north';
    case 'south':
      return 'rsu-south';
    case 'east':
      return 'rsu-east';
    case 'west':
      return 'rsu-west';
  }
}

function createSessionKeyRef(vehicleId: string) {
  return `sess-${vehicleId.slice(-2)}a7`;
}

function deriveCommunicationState({
  vehicleId,
  path,
  wrappedProgress,
  lightColor,
  moving,
}: {
  vehicleId: string;
  path: CarProps['path'];
  wrappedProgress: number;
  lightColor: 'red' | 'yellow' | 'green';
  moving: boolean;
}) {
  const rsuId = rsuIdForPath(path);
  const sessionKeyRef = createSessionKeyRef(vehicleId);

  if (!moving && wrappedProgress > 13 && wrappedProgress < 26) {
    return {
      stage: 'waiting' as const,
      message: lightColor === 'red' ? 'Holding secure session at red light' : 'Preparing green-light release',
      sessionKeyRef,
      rsuId,
      kemAlgorithm: 'ML-KEM-768',
      signatureAlgorithm: 'ML-DSA',
      sessionCipher: 'AES-256-GCM',
      bubbleVisible: true,
    };
  }

  if (wrappedProgress < 10) {
    return {
      stage: 'idle' as const,
      message: 'Scanning V2X edge coverage',
      sessionKeyRef,
      rsuId,
      kemAlgorithm: 'ML-KEM-768',
      signatureAlgorithm: 'ML-DSA',
      sessionCipher: 'AES-256-GCM',
      bubbleVisible: false,
    };
  }

  if (wrappedProgress < 16) {
    return {
      stage: 'beacon' as const,
      message: `Beacon verified from ${rsuId}`,
      sessionKeyRef,
      rsuId,
      kemAlgorithm: 'ML-KEM-768',
      signatureAlgorithm: 'ML-DSA',
      sessionCipher: 'AES-256-GCM',
      bubbleVisible: true,
    };
  }

  if (wrappedProgress < 22) {
    return {
      stage: 'auth' as const,
      message: 'Vehicle certificate checked with ML-DSA',
      sessionKeyRef,
      rsuId,
      kemAlgorithm: 'ML-KEM-768',
      signatureAlgorithm: 'ML-DSA',
      sessionCipher: 'AES-256-GCM',
      bubbleVisible: true,
    };
  }

  if (wrappedProgress < 30) {
    return {
      stage: 'join' as const,
      message: 'Vehicle_Join_Request transmitted',
      sessionKeyRef,
      rsuId,
      kemAlgorithm: 'ML-KEM-768',
      signatureAlgorithm: 'ML-DSA',
      sessionCipher: 'AES-256-GCM',
      bubbleVisible: true,
    };
  }

  if (wrappedProgress < 40) {
    return {
      stage: 'secure' as const,
      message: 'AES-256-GCM telemetry tunnel active',
      sessionKeyRef,
      rsuId,
      kemAlgorithm: 'ML-KEM-768',
      signatureAlgorithm: 'ML-DSA',
      sessionCipher: 'AES-256-GCM',
      bubbleVisible: true,
    };
  }

  return {
    stage: 'secure' as const,
    message: 'Session keepalive with roadside edge',
    sessionKeyRef,
    rsuId,
    kemAlgorithm: 'ML-KEM-768',
    signatureAlgorithm: 'ML-DSA',
    sessionCipher: 'AES-256-GCM',
    bubbleVisible: false,
  };
}
