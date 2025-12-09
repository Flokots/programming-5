import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet, Dimensions } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { GameState, ColorOption } from '../types';

interface GameScreenProps {
  gameState: GameState;
  username: string;
  answered: boolean;
  showWrongAnswer: boolean;
  onColorClick: (color: ColorOption) => void;
}

const { width } = Dimensions.get('window');
const buttonSize = (width - 80) / 2;

export default function GameScreen({
  gameState,
  username,
  answered,
  showWrongAnswer,
  onColorClick,
}: GameScreenProps) {
  const colorButtons: Array<{ color: ColorOption; label: string; bg: string }> = [
    { color: 'red', label: 'RED', bg: '#ff4444' },
    { color: 'blue', label: 'BLUE', bg: '#4444ff' },
    { color: 'green', label: 'GREEN', bg: '#44ff44' },
    { color: 'yellow', label: 'YELLOW', bg: '#ffff44' },
  ];

  return (
    <SafeAreaView style={styles.container}>
      {/* Header with Scores */}
      <View style={styles.header}>
        <View style={styles.playerInfo}>
          <Text style={styles.playerName}>{username}</Text>
          <Text style={styles.score}>{gameState.yourScore}</Text>
        </View>

        <View style={styles.roundInfo}>
          <Text style={styles.roundText}>ROUND</Text>
          <Text style={styles.roundNumber}>
            {gameState.currentRound}/{gameState.maxRounds}
          </Text>
        </View>

        <View style={styles.playerInfo}>
          <Text style={styles.playerName}>Opponent</Text>
          <Text style={styles.score}>{gameState.opponentScore}</Text>
        </View>
      </View>

      {/* Game Content */}
      {gameState.currentRound === 0 ? (
        <View style={styles.instructions}>
          <Text style={styles.instructionsEmoji}>🎮</Text>
          <Text style={styles.instructionsTitle}>GAME STARTING!</Text>
          <Text style={styles.instructionsText}>
            Match the <Text style={styles.bold}>COLOR</Text> of the text,{'\n'}
            not the word!
          </Text>
          <Text style={styles.instructionsRounds}>
            You will play {gameState.maxRounds} rounds
          </Text>
        </View>
      ) : (
        <>
          {/* Word Display */}
          <View style={styles.wordContainer}>
            <Text style={styles.prompt}>What COLOR is this text?</Text>
            <Text style={[styles.word, { color: gameState.color }]}>
              {gameState.word}
            </Text>
          </View>

          {/* Color Buttons */}
          <View style={styles.buttonsContainer}>
            <View style={styles.buttonRow}>
              {colorButtons.slice(0, 2).map((btn) => (
                <TouchableOpacity
                  key={btn.color}
                  style={[
                    styles.colorButton,
                    { backgroundColor: btn.bg },
                    answered && styles.buttonDisabled,
                  ]}
                  onPress={() => onColorClick(btn.color)}
                  disabled={answered}
                  activeOpacity={0.7}
                >
                  <Text style={styles.buttonLabel}>{btn.label}</Text>
                </TouchableOpacity>
              ))}
            </View>
            <View style={styles.buttonRow}>
              {colorButtons.slice(2, 4).map((btn) => (
                <TouchableOpacity
                  key={btn.color}
                  style={[
                    styles.colorButton,
                    { backgroundColor: btn.bg },
                    answered && styles.buttonDisabled,
                  ]}
                  onPress={() => onColorClick(btn.color)}
                  disabled={answered}
                  activeOpacity={0.7}
                >
                  <Text style={styles.buttonLabel}>{btn.label}</Text>
                </TouchableOpacity>
              ))}
            </View>
          </View>

          {/* Feedback */}
          {showWrongAnswer && (
            <View style={styles.feedbackWrong}>
              <Text style={styles.feedbackText}>❌ Wrong! Try again!</Text>
            </View>
          )}

          {answered && !showWrongAnswer && (
            <View style={styles.feedbackCorrect}>
              <Text style={styles.feedbackText}>✅ Answer submitted!</Text>
            </View>
          )}
        </>
      )}
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1a1a2e',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingVertical: 16,
    backgroundColor: '#16213e',
    borderBottomWidth: 2,
    borderBottomColor: '#0f3460',
  },
  playerInfo: {
    alignItems: 'center',
  },
  playerName: {
    fontSize: 14,
    color: '#aaa',
    marginBottom: 4,
  },
  score: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#e94560',
  },
  roundInfo: {
    alignItems: 'center',
  },
  roundText: {
    fontSize: 12,
    color: '#aaa',
    letterSpacing: 1,
  },
  roundNumber: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
  },
  instructions: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  instructionsEmoji: {
    fontSize: 80,
    marginBottom: 24,
  },
  instructionsTitle: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#e94560',
    marginBottom: 24,
    letterSpacing: 2,
  },
  instructionsText: {
    fontSize: 20,
    color: '#fff',
    textAlign: 'center',
    lineHeight: 32,
    marginBottom: 16,
  },
  bold: {
    fontWeight: 'bold',
    color: '#e94560',
  },
  instructionsRounds: {
    fontSize: 16,
    color: '#aaa',
  },
  wordContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 24,
  },
  prompt: {
    fontSize: 18,
    color: '#aaa',
    marginBottom: 24,
    textAlign: 'center',
  },
  word: {
    fontSize: 72,
    fontWeight: 'bold',
    textAlign: 'center',
  },
  buttonsContainer: {
    padding: 20,
    paddingBottom: 40,
  },
  buttonRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 16,
  },
  colorButton: {
    width: buttonSize,
    height: buttonSize,
    borderRadius: 16,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 4,
    borderColor: '#000',
    elevation: 8,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
  },
  buttonDisabled: {
    opacity: 0.5,
  },
  buttonLabel: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#000',
    textShadowColor: 'rgba(255, 255, 255, 0.3)',
    textShadowOffset: { width: 1, height: 1 },
    textShadowRadius: 2,
  },
  feedbackWrong: {
    position: 'absolute',
    top: '50%',
    left: 24,
    right: 24,
    backgroundColor: '#ff4444',
    padding: 20,
    borderRadius: 12,
    alignItems: 'center',
  },
  feedbackCorrect: {
    position: 'absolute',
    top: '50%',
    left: 24,
    right: 24,
    backgroundColor: '#44ff44',
    padding: 20,
    borderRadius: 12,
    alignItems: 'center',
  },
  feedbackText: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#000',
  },
});