import React, { useState, useEffect } from "react";

export const Header: React.FC = () => {
  const [time, setTime] = useState<Date>(new Date());

  useEffect(() => {
    const timerId = setInterval(() => {
      setTime(new Date());
    }, 1000);

    // Cleanup the interval on component unmount
    return () => clearInterval(timerId);
  }, []);

  return (
    <header className="flex w-full flex-col items-center justify-center gap-0.5 py-2 sm:py-10">
      <p className="text-[18px] font-medium">{time.toLocaleTimeString()}</p>
      <p>
        {time.toLocaleDateString("en-GB", {
          day: "numeric",
          month: "short",
          year: "numeric",
        })}
      </p>
    </header>
  );
};
