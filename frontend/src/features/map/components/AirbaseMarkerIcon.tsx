import type { CSSProperties } from 'react';
import { memo } from 'react';
import { PiTriangleDuotone, PiTriangleFill } from 'react-icons/pi';

export type AirbaseMarkerVariant = 'default' | 'hovered' | 'selected';

type AirbaseMarkerIconProps = {
  variant: AirbaseMarkerVariant;
  sizePx: number;
};

type MarkerVisual = {
  color: string;
  haloStyle?: CSSProperties;
  Icon: typeof PiTriangleDuotone;
};

function resolveMarkerVisual(variant: AirbaseMarkerVariant): MarkerVisual {
  switch (variant) {
    case 'selected':
      return {
        Icon: PiTriangleFill,
        color: 'var(--color-airbase-selected-fill)',
        haloStyle: {
          backgroundColor: 'rgba(255, 255, 255, 0.12)',
          boxShadow: 'inset 0 0 0 2px var(--color-airbase-selected-border)',
        },
      };
    case 'hovered':
      return {
        Icon: PiTriangleDuotone,
        color: 'var(--color-airbase-hover)',
        haloStyle: {
          backgroundColor: 'rgba(255, 255, 255, 0.08)',
          boxShadow: 'inset 0 0 0 1px var(--color-airbase-hover)',
        },
      };
    default:
      return {
        Icon: PiTriangleDuotone,
        color: 'var(--color-airbase-default-fill)',
      };
  }
}

function AirbaseMarkerIconComponent({
  variant,
  sizePx,
}: AirbaseMarkerIconProps) {
  const { Icon, color, haloStyle } = resolveMarkerVisual(variant);
  const haloSize = sizePx + 10;

  return (
    <>
      <span
        aria-hidden="true"
        className="absolute rounded-full transition-all"
        style={{
          width: haloSize,
          height: haloSize,
          ...haloStyle,
        }}
      />
      <Icon
        aria-hidden="true"
        size={sizePx}
        className="relative shrink-0"
        style={{
          color,
        }}
      />
    </>
  );
}

export const AirbaseMarkerIcon = memo(AirbaseMarkerIconComponent);
