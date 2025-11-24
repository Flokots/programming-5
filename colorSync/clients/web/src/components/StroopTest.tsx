import { useState } from "react";
import { type StroopColor } from "../types/game";
import "./StroopTest.css";

interface StroopTestProps {
  onColorSelect: (color: StroopColor) => void;
  disabled: boolean;
}

const COLORS: StroopColor[] = ["red", "blue", "green", "yellow"];

export default function StroopTest({
  onColorSelect,
  disabled,
}: StroopTestProps) {
  const [clickedColor, setClickedColor] = useState<StroopColor | null>(null);

  const handleClick = (color: StroopColor) => {
    if (disabled) return;

    // Set clicked state for animation
    setClickedColor(color);

    // Call parent handler
    onColorSelect(color);

    // Reset clicked state after animation duration
    setTimeout(() => {
      setClickedColor(null);
    }, 300);
  };

  return (
    <div className="stroop-test">
      <div className="color-buttons">
        {COLORS.map((color) => (
          <button
            key={color}
            className={`color-button color-${color} ${
              clickedColor === color ? "clicked" : ""
            } 
            ${disabled ? "disabled" : ""}`}
            onClick={() => handleClick(color)}
            disabled={disabled}
          >
            {color.toUpperCase()}
          </button>
        ))}
      </div>

      {disabled && (
        <div className="waiting-indicator">
          <div className="spinner"></div>
          <p>Waiting for next round...</p>
        </div>
      )}
    </div>
  );
}
