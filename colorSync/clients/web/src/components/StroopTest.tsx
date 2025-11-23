import { COLOR_MAP, type StroopColor } from '../types/game';
import './StroopTest.css';

interface StroopTestProps {
  word: string;
  color: string;
}

function StroopTest({ word, color }: StroopTestProps) {
  // Get the hex color value from our color map
  const colorValue = COLOR_MAP[color as StroopColor] || '#000000';

  return (
    <div className="stroop-test">
      <div className="stroop-instruction">
        Click the button that matches the <strong>COLOR</strong> (not the word)
      </div>
      
      <div 
        className="stroop-word"
        style={{ color: colorValue }}
      >
        {word.toUpperCase()}
      </div>

      <div className="stroop-hint">
        The word says "{word}" but what COLOR is it?
      </div>
    </div>
  );
}

export default StroopTest;