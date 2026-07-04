import { brandIcons, type BrandIconName } from "../../assets/brand";

interface BrandIconProps {
  name: BrandIconName;
  alt?: string;
  size?: number;
  className?: string;
}

export default function BrandIcon({ name, alt = "", size = 24, className = "" }: BrandIconProps) {
  return (
    <img
      src={brandIcons[name]}
      alt={alt}
      width={size}
      height={size}
      className={`inline-block object-contain ${className}`}
      aria-hidden={alt ? undefined : true}
    />
  );
}
