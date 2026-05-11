import { useRef, useMemo, useEffect } from 'react';
import { useFrame } from '@react-three/fiber';
import { Box } from '@react-three/drei';
import * as THREE from 'three';
import { getLightColorForPath } from '@/stores/trafficStore';
import { registerCar, updateCar, unregisterCar, getCarAhead } from '@/stores/carStore';

interface CarProps {
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

export function Car({ path, lane, speed: baseSpeed = 4, color = '#ef4444', startOffset = 0 }: CarProps) {
  const ref = useRef<THREE.Group>(null);
  const progress = useRef(startOffset);
  const currentSpeed = useRef(baseSpeed);
  const carId = useRef(`car-${Math.random().toString(36).slice(2, 9)}`);

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
    return () => unregisterCar(id);
  }, [path, lane]);

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
  });

  return (
    <group ref={ref} position={startPos.toArray()} castShadow>
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
