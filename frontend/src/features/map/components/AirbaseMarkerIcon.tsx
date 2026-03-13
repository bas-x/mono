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
          backgroundColor: 'color-mix(in srgb, var(--color-airbase-selected-fill) 22%, transparent)',
          boxShadow:
            '0 0 18px color-mix(in srgb, var(--color-airbase-selected-fill) 42%, transparent), inset 0 0 0 2px var(--color-airbase-selected-border)',
        },
      };
    case 'hovered':
      return {
        Icon: PiTriangleDuotone,
        color: 'var(--color-airbase-hover)',
        haloStyle: {
          backgroundColor: 'color-mix(in srgb, var(--color-airbase-hover) 18%, transparent)',
          boxShadow:
            '0 0 16px color-mix(in srgb, var(--color-airbase-hover) 28%, transparent), inset 0 0 0 1px color-mix(in srgb, var(--color-airbase-hover) 70%, var(--color-airbase-default-stroke) 30%)',
        },
      };
    default:
      return {
        Icon: PiTriangleDuotone,
        color: 'var(--color-airbase-default-fill)',
        haloStyle: {
          backgroundColor:
            'color-mix(in srgb, var(--color-airbase-default-fill) 10%, transparent)',
          boxShadow:
            'inset 0 0 0 1px color-mix(in srgb, var(--color-airbase-default-stroke) 55%, transparent)',
        },
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
          filter:
            variant === 'selected'
              ? 'drop-shadow(0 0 8px color-mix(in srgb, var(--color-airbase-selected-fill) 40%, transparent))'
              : 'drop-shadow(0 0 6px color-mix(in srgb, var(--color-airbase-default-stroke) 28%, transparent))',
        }}
      />
    </>
  );
}

export const AirbaseMarkerIcon = memo(AirbaseMarkerIconComponent);
