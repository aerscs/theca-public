export const Loader = (props: React.SVGProps<SVGSVGElement>) => (
  <svg
    xmlns="http://www.w3.org/2000/svg"
    width={24}
    height={24}
    viewBox="0 0 24 24"
    className="animate-apple"
    fill="none"
    {...props}
  >
    <path
      d="M12 2V6"
      stroke="white"
      strokeWidth={props.strokeWidth || 2}
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      opacity={0}
      d="M16.2 7.79999L19.1 4.89999"
      stroke="white"
      strokeWidth={props.strokeWidth || 2}
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      opacity={0.15}
      d="M18 12H22"
      stroke="white"
      strokeWidth={props.strokeWidth || 2}
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      opacity={0.3}
      d="M16.2 16.2L19.1 19.1"
      stroke="white"
      strokeWidth={props.strokeWidth || 2}
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      opacity={0.45}
      d="M12 18V22"
      stroke="white"
      strokeWidth={props.strokeWidth || 2}
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      opacity={0.6}
      d="M4.90002 19.1L7.80002 16.2"
      stroke="white"
      strokeWidth={props.strokeWidth || 2}
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      opacity={0.75}
      d="M2 12H6"
      stroke="white"
      strokeWidth={props.strokeWidth || 2}
      strokeLinecap="round"
      strokeLinejoin="round"
    />
    <path
      opacity={0.9}
      d="M4.90002 4.89999L7.80002 7.79999"
      stroke="white"
      strokeWidth={props.strokeWidth || 2}
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
);
