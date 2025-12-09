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

const { width, height } = Dimensions.get('window');
const isSmallScreen = height < 700;
const buttonSize = Math.min((width - 64) / 2, 140); // Max 140px, with proper spacing

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
    <SafeAreaView style={styles.container} edges={['top']}>
      {/* Header with Scores */}
      <View style={styles.header}>
        <View style={styles.playerInfo}>
          <Text style={styles.playerName} numberOfLines={1}>{username}</Text>
          <Text style={styles.score}>{gameState.yourScore}</Text>
        </View>

        <View style={styles.roundInfo}>
          <Text style={styles.roundText}>ROUND</Text>
          <Text style={styles.roundNumber}>
            {gameState.currentRound}/{gameState.maxRounds}
          </Text>
        </View>

        <View style={styles.playerInfo}>
          <Text style={styles.playerName} numberOfLines={1}>Opponent</Text>
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
        <View style={styles.gameContent}>
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
                    { backgroundColor: btn.bg, width: buttonSize, height: buttonSize },
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
                    { backgroundColor: btn.bg, width: buttonSize, height: buttonSize },
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
        </View>
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
    paddingHorizontal: 16,
    paddingVertical: 12,
    backgroundColor: '#16213e',
    borderBottomWidth: 2,
    borderBottomColor: '#0f3460',
  },
  playerInfo: {
    alignItems: 'center',
    flex: 1,
  },
  playerName: {
    fontSize: 12,
    color: '#aaa',
    marginBottom: 2,
  },
  score: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#e94560',
  },
  roundInfo: {
    alignItems: 'center',
    flex: 1,
  },
  roundText: {
    fontSize: 10,
    color: '#aaa',
    letterSpacing: 1,
  },
  roundNumber: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#fff',
  },
  instructions: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },
  instructionsEmoji: {
    fontSize: 64,
    marginBottom: 20,
  },
  instructionsTitle: {
    fontSize: 28,
    fontWeight: 'bold',
    color: '#e94560',
    marginBottom: 20,
    letterSpacing: 2,
  },
  instructionsText: {
    fontSize: 18,
    color: '#fff',
    textAlign: 'center',
    lineHeight: 28,
    marginBottom: 12,
  },
  bold: {
    fontWeight: 'bold',
    color: '#e94560',
  },
  instructionsRounds: {
    fontSize: 14,
    color: '#aaa',
  },
  gameContent: {
    flex: 1,
    justifyContent: 'space-between',
  },
  wordContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 24,
    paddingVertical: 20,
  },
  prompt: {
    fontSize: isSmallScreen ? 14 : 16,
    color: '#aaa',
    marginBottom: 16,
    textAlign: 'center',
  },
  word: {
    fontSize: isSmallScreen ? 48 : 64,
    fontWeight: 'bold',
    textAlign: 'center',
  },
  buttonsContainer: {
    paddingHorizontal: 20,
    paddingBottom: 20,
  },
  buttonRow: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    marginBottom: 12,
  },
  colorButton: {
    borderRadius: 12,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 3,
    borderColor: '#000',
    elevation: 6,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 3 },
    shadowOpacity: 0.3,
    shadowRadius: 6,
  },
  buttonDisabled: {
    opacity: 0.5,
  },
  buttonLabel: {
    fontSize: isSmallScreen ? 18 : 22,
    fontWeight: 'bold',
    color: '#000',
    textShadowColor: 'rgba(255, 255, 255, 0.3)',
    textShadowOffset: { width: 1, height: 1 },
    textShadowRadius: 2,
  },
  feedbackWrong: {
    position: 'absolute',
    top: '45%',
    left: 24,
    right: 24,
    backgroundColor: '#ff4444',
    padding: 16,
    borderRadius: 12,
    alignItems: 'center',
    elevation: 10,
  },
  feedbackCorrect: {
    position: 'absolute',
    top: '45%',
    left: 24,
    right: 24,
    backgroundColor: '#44ff44',
    padding: 16,
    borderRadius: 12,
    alignItems: 'center',
    elevation: 10,
  },
  feedbackText: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#000',
  },
});