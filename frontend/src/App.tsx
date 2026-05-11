import { useState, useEffect } from 'react';
import { Scene } from '@/components/Scene';
import { getPhaseLabel, getTimeRemaining, trafficState, type LightColor } from '@/stores/trafficStore';
import { IconInfoCircle, IconMap, IconWifi, IconTrafficLights } from '@tabler/icons-react';

function LightIndicator({ color, label, path }: { color: LightColor; label: string; path: 'north' | 'south' | 'east' | 'west' }) {
  const colorClasses: Record<LightColor, string> = {
    red: 'bg-red-500 shadow-red-500/50',
    yellow: 'bg-yellow-400 shadow-yellow-400/50',
    green: 'bg-green-500 shadow-green-500/50',
  };

  const dimClasses: Record<LightColor, string> = {
    red: 'bg-red-900/40',
    yellow: 'bg-yellow-900/40',
    green: 'bg-green-900/40',
  };

  const isActive = trafficState[path] === color;

  return (
    <div className="flex items-center gap-1.5">
      <span className="text-[10px] w-8 text-slate-400 uppercase tracking-wide">{label}</span>
      <div
        className={`h-2.5 w-2.5 rounded-full transition-all duration-300 ${
          isActive ? colorClasses[color] : dimClasses[color]
        } ${isActive ? 'shadow-[0_0_6px]' : ''}`}
      />
    </div>
  );
}

function TrafficStatusPanel() {
  const [, setTick] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => setTick((t) => t + 1), 100);
    return () => clearInterval(interval);
  }, []);

  const phase = getPhaseLabel();
  const remaining = getTimeRemaining();

  return (
    <div className="rounded-2xl border border-white/10 bg-slate-950/70 p-4 shadow-xl backdrop-blur-md">
      <div className="mb-3 flex items-center gap-2">
        <IconTrafficLights className="h-4 w-4 text-amber-400" />
        <h2 className="text-sm font-semibold text-white">Traffic Control</h2>
      </div>

      <div className="mb-3 flex items-center justify-between">
        <span className="text-xs text-slate-400">Phase</span>
        <span className="rounded-lg bg-slate-800/80 px-2 py-0.5 text-xs font-medium text-white">
          {phase}
        </span>
      </div>

      <div className="mb-3 flex items-center justify-between">
        <span className="text-xs text-slate-400">Remaining</span>
        <span className="text-xs font-mono text-slate-300">{remaining.toFixed(1)}s</span>
      </div>

      <div className="space-y-1.5 border-t border-white/5 pt-3">
        <div className="flex items-center justify-between">
          <span className="text-[10px] text-slate-500 uppercase tracking-wide">NS Road</span>
          <div className="flex gap-2">
            <LightIndicator color={trafficState.north} label="N" path="north" />
            <LightIndicator color={trafficState.south} label="S" path="south" />
          </div>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-[10px] text-slate-500 uppercase tracking-wide">EW Road</span>
          <div className="flex gap-2">
            <LightIndicator color={trafficState.east} label="E" path="east" />
            <LightIndicator color={trafficState.west} label="W" path="west" />
          </div>
        </div>
      </div>
    </div>
  );
}

function App() {
  const [showInfo, setShowInfo] = useState(true);

  return (
    <div className="relative h-screen w-screen overflow-hidden bg-slate-900">
      <Scene />

      {/* Header overlay */}
      <div className="pointer-events-none absolute top-0 left-0 right-0 flex items-start justify-between p-5">
        <div className="pointer-events-auto rounded-2xl border border-white/10 bg-slate-950/70 px-5 py-4 shadow-xl backdrop-blur-md">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-emerald-500/15">
              <IconMap className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <h1 className="text-base font-semibold tracking-tight text-white">
                AVMS Simulation
              </h1>
              <p className="text-xs text-slate-400">
                Isometric V2X junction view
              </p>
            </div>
          </div>
        </div>

        <button
          onClick={() => setShowInfo((s) => !s)}
          className="pointer-events-auto flex h-10 w-10 items-center justify-center rounded-xl border border-white/10 bg-slate-950/70 text-slate-300 shadow-lg backdrop-blur-md transition-colors hover:bg-slate-900/90 hover:text-white"
          aria-label="Toggle info"
        >
          <IconInfoCircle className="h-5 w-5" />
        </button>
      </div>

      {/* Info panels */}
      {showInfo && (
        <div className="pointer-events-none absolute right-5 bottom-5 flex max-w-xs flex-col gap-3">
          <TrafficStatusPanel />

          <div className="pointer-events-auto rounded-2xl border border-white/10 bg-slate-950/70 p-4 shadow-xl backdrop-blur-md">
            <div className="mb-2 flex items-center gap-2">
              <IconWifi className="h-4 w-4 text-sky-400" />
              <h2 className="text-sm font-semibold text-white">RSU Coverage</h2>
            </div>
            <p className="text-xs leading-relaxed text-slate-400">
              4 Roadside Units (RSU) are positioned at the junction corners.
              The blue pulses indicate active V2X broadcast range.
            </p>
            <div className="mt-2 flex flex-wrap gap-2">
              <span className="rounded-lg bg-emerald-500/10 px-2 py-1 text-[10px] font-medium text-emerald-400">
                8 Vehicles
              </span>
              <span className="rounded-lg bg-sky-500/10 px-2 py-1 text-[10px] font-medium text-sky-400">
                4 RSUs
              </span>
              <span className="rounded-lg bg-amber-500/10 px-2 py-1 text-[10px] font-medium text-amber-400">
                4 Traffic Lights
              </span>
            </div>
            <p className="mt-2 text-[10px] text-slate-500">
              Drag to pan • Scroll to zoom
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;
