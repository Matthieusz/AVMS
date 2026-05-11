import { useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import { Box, Circle, Cylinder } from '@react-three/drei';
import * as THREE from 'three';
import { trafficState, type LightColor } from '@/stores/trafficStore';

interface TrafficLightProps {
  position: [number, number, number];
  rotation: [number, number, number];
  path: 'north' | 'south' | 'east' | 'west';
}

const activeColors: Record<LightColor, string> = {
  red: '#ef4444',
  yellow: '#eab308',
  green: '#22c55e',
};

const dimColors: Record<LightColor, string> = {
  red: '#450a0a',
  yellow: '#422006',
  green: '#052e16',
};

function LightBulb({
  lightColor,
  position,
  path,
}: {
  lightColor: LightColor;
  position: [number, number, number];
  path: 'north' | 'south' | 'east' | 'west';
}) {
  const meshRef = useRef<THREE.Mesh>(null);
  const glowRef = useRef<THREE.Mesh>(null);

  useFrame(() => {
    if (!meshRef.current) return;
    const currentColor = trafficState[path];
    const isActive = currentColor === lightColor;

    const mat = meshRef.current.material as THREE.MeshStandardMaterial;
    mat.color.set(isActive ? activeColors[lightColor] : dimColors[lightColor]);
    mat.emissive.set(isActive ? activeColors[lightColor] : '#000000');
    mat.emissiveIntensity = isActive ? 2.5 : 0;

    if (glowRef.current) {
      const glowMat = glowRef.current.material as THREE.MeshStandardMaterial;
      glowMat.opacity = isActive ? 0.2 + Math.sin(Date.now() * 0.004) * 0.08 : 0;
    }
  });

  return (
    <group position={position}>
      <Circle ref={meshRef} args={[0.1, 32]}>
        <meshStandardMaterial side={THREE.DoubleSide} />
      </Circle>
      <Circle ref={glowRef} args={[0.2, 32]} position={[0, 0, -0.02]}>
        <meshStandardMaterial color={activeColors[lightColor]} transparent opacity={0} side={THREE.DoubleSide} />
      </Circle>
      {/* Visor */}
      <Box args={[0.24, 0.06, 0.1]} position={[0, 0.09, 0.04]}>
        <meshStandardMaterial color="#1f2937" />
      </Box>
    </group>
  );
}

export function TrafficLight({ position, rotation, path }: TrafficLightProps) {
  return (
    <group position={position} rotation={rotation}>
      {/* Pole */}
      <Cylinder args={[0.08, 0.1, 3.5]} position={[0, 1.75, 0]} castShadow>
        <meshStandardMaterial color="#525252" metalness={0.5} roughness={0.5} />
      </Cylinder>

      {/* Light housing */}
      <Box args={[0.35, 1.05, 0.2]} position={[0, 3.85, 0]} castShadow>
        <meshStandardMaterial color="#1f2937" roughness={0.3} />
      </Box>

      {/* Backing plate */}
      <Box args={[0.3, 1.0, 0.05]} position={[0, 3.85, -0.05]}>
        <meshStandardMaterial color="#111827" />
      </Box>

      {/* Red light - top */}
      <LightBulb lightColor="red" path={path} position={[0, 4.2, 0.11]} />
      {/* Yellow light - middle */}
      <LightBulb lightColor="yellow" path={path} position={[0, 3.85, 0.11]} />
      {/* Green light - bottom */}
      <LightBulb lightColor="green" path={path} position={[0, 3.5, 0.11]} />
    </group>
  );
}
