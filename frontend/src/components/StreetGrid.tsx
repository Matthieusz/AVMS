import { Box, Plane } from '@react-three/drei';

const ROAD_WIDTH = 6;
const ROAD_LENGTH = 60;
const SIDEWALK_WIDTH = 2;
const SIDEWALK_HEIGHT = 0.15;
const STOP_LINE_OFFSET = 5;

function RoadSegment({
  position,
  rotation = [0, 0, 0],
  length = ROAD_LENGTH,
}: {
  position: [number, number, number];
  rotation?: [number, number, number];
  length?: number;
}) {
  return (
    <group position={position} rotation={rotation}>
      {/* Asphalt */}
      <Plane args={[ROAD_WIDTH, length]} rotation={[-Math.PI / 2, 0, 0]} receiveShadow>
        <meshStandardMaterial color="#3a3a3a" roughness={0.9} />
      </Plane>

      {/* Center dashed line */}
      {Array.from({ length: Math.floor(length / 3) }).map((_, i) => (
        <Plane
          key={i}
          args={[0.25, 1.2]}
          position={[0, 0.01, -length / 2 + 1.5 + i * 3]}
          rotation={[-Math.PI / 2, 0, 0]}
        >
          <meshStandardMaterial color="#e8e8e8" />
        </Plane>
      ))}

      {/* Edge lines */}
      <Plane
        args={[0.15, length]}
        position={[-ROAD_WIDTH / 2 + 0.3, 0.01, 0]}
        rotation={[-Math.PI / 2, 0, 0]}
      >
        <meshStandardMaterial color="#e8e8e8" />
      </Plane>
      <Plane
        args={[0.15, length]}
        position={[ROAD_WIDTH / 2 - 0.3, 0.01, 0]}
        rotation={[-Math.PI / 2, 0, 0]}
      >
        <meshStandardMaterial color="#e8e8e8" />
      </Plane>
    </group>
  );
}

function Sidewalk({
  position,
  rotation = [0, 0, 0],
  width = SIDEWALK_WIDTH,
  length = ROAD_LENGTH,
}: {
  position: [number, number, number];
  rotation?: [number, number, number];
  width?: number;
  length?: number;
}) {
  return (
    <Box args={[width, SIDEWALK_HEIGHT, length]} position={position} rotation={rotation} castShadow receiveShadow>
      <meshStandardMaterial color="#9ca3af" roughness={0.8} />
    </Box>
  );
}

function Crosswalk({ position, rotation = [0, 0, 0] }: { position: [number, number, number]; rotation?: [number, number, number] }) {
  return (
    <group position={position} rotation={rotation}>
      {Array.from({ length: 6 }).map((_, i) => (
        <Plane
          key={i}
          args={[0.8, ROAD_WIDTH - 0.5]}
          position={[-2 + i * 0.85, 0.015, 0]}
          rotation={[-Math.PI / 2, 0, 0]}
        >
          <meshStandardMaterial color="#f0f0f0" />
        </Plane>
      ))}
    </group>
  );
}

function StopLine({ position, args }: { position: [number, number, number]; args: [number, number] }) {
  return (
    <Plane args={args} position={position} rotation={[-Math.PI / 2, 0, 0]}>
      <meshStandardMaterial color="#f0f0f0" />
    </Plane>
  );
}

export function StreetGrid() {
  const sw = ROAD_WIDTH / 2 + SIDEWALK_WIDTH / 2;

  return (
    <group>
      {/* Roads */}
      <RoadSegment position={[0, 0, 0]} rotation={[0, 0, 0]} />
      <RoadSegment position={[0, 0, 0]} rotation={[0, Math.PI / 2, 0]} />

      {/* Sidewalks - North */}
      <Sidewalk position={[-sw, SIDEWALK_HEIGHT / 2, -ROAD_LENGTH / 2 - ROAD_WIDTH / 2]} length={ROAD_LENGTH} />
      <Sidewalk position={[sw, SIDEWALK_HEIGHT / 2, -ROAD_LENGTH / 2 - ROAD_WIDTH / 2]} length={ROAD_LENGTH} />

      {/* Sidewalks - South */}
      <Sidewalk position={[-sw, SIDEWALK_HEIGHT / 2, ROAD_LENGTH / 2 + ROAD_WIDTH / 2]} length={ROAD_LENGTH} />
      <Sidewalk position={[sw, SIDEWALK_HEIGHT / 2, ROAD_LENGTH / 2 + ROAD_WIDTH / 2]} length={ROAD_LENGTH} />

      {/* Sidewalks - East */}
      <Sidewalk position={[ROAD_LENGTH / 2 + ROAD_WIDTH / 2, SIDEWALK_HEIGHT / 2, -sw]} rotation={[0, Math.PI / 2, 0]} length={ROAD_LENGTH} />
      <Sidewalk position={[ROAD_LENGTH / 2 + ROAD_WIDTH / 2, SIDEWALK_HEIGHT / 2, sw]} rotation={[0, Math.PI / 2, 0]} length={ROAD_LENGTH} />

      {/* Sidewalks - West */}
      <Sidewalk position={[-ROAD_LENGTH / 2 - ROAD_WIDTH / 2, SIDEWALK_HEIGHT / 2, -sw]} rotation={[0, Math.PI / 2, 0]} length={ROAD_LENGTH} />
      <Sidewalk position={[-ROAD_LENGTH / 2 - ROAD_WIDTH / 2, SIDEWALK_HEIGHT / 2, sw]} rotation={[0, Math.PI / 2, 0]} length={ROAD_LENGTH} />

      {/* Corner sidewalk fills */}
      <Sidewalk position={[-sw, SIDEWALK_HEIGHT / 2, -sw]} width={SIDEWALK_WIDTH} length={SIDEWALK_WIDTH} />
      <Sidewalk position={[sw, SIDEWALK_HEIGHT / 2, -sw]} width={SIDEWALK_WIDTH} length={SIDEWALK_WIDTH} />
      <Sidewalk position={[-sw, SIDEWALK_HEIGHT / 2, sw]} width={SIDEWALK_WIDTH} length={SIDEWALK_WIDTH} />
      <Sidewalk position={[sw, SIDEWALK_HEIGHT / 2, sw]} width={SIDEWALK_WIDTH} length={SIDEWALK_WIDTH} />

      {/* Stop lines at each approach */}
      <StopLine position={[0, 0.02, STOP_LINE_OFFSET]} args={[ROAD_WIDTH - 0.5, 0.25]} />
      <StopLine position={[0, 0.02, -STOP_LINE_OFFSET]} args={[ROAD_WIDTH - 0.5, 0.25]} />
      <StopLine position={[-STOP_LINE_OFFSET, 0.02, 0]} args={[0.25, ROAD_WIDTH - 0.5]} />
      <StopLine position={[STOP_LINE_OFFSET, 0.02, 0]} args={[0.25, ROAD_WIDTH - 0.5]} />

      {/* Crosswalks at each approach */}
      <Crosswalk position={[-ROAD_WIDTH / 2 - 2.5, 0, 0]} rotation={[0, 0, Math.PI / 2]} />
      <Crosswalk position={[ROAD_WIDTH / 2 + 2.5, 0, 0]} rotation={[0, 0, Math.PI / 2]} />
      <Crosswalk position={[0, 0, -ROAD_WIDTH / 2 - 2.5]} />
      <Crosswalk position={[0, 0, ROAD_WIDTH / 2 + 2.5]} />
    </group>
  );
}
