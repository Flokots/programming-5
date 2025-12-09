import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet, ScrollView } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import type { GameState } from '../types';

interface GameOverScreenProps {
  gameState: GameState;
  userId: string;
  onPlayAgain: () => void;
}

export default function GameOverScreen({ gameState, userId, onPlayAgain }: GameOverScreenProps) {
  const isWinner = gameState.winner === userId;
  const isDraw = gameState.winner === 'draw';

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView contentContainerStyle={styles.scrollContent}>
        {/* Header */}
        <View style={styles.header}>
          <Text style={styles.title}>🏁 GAME OVER</Text>
        </View>

        {/* Result Banner */}
        <View style={[
          styles.banner,
          isWinner && styles.bannerWin,
          isDraw && styles.bannerDraw,
          !isWinner && !isDraw && styles.bannerLose,
        ]}>
          <Text style={styles.bannerText}>
            {isWinner && '🎉 YOU WON! 🎉'}
            {isDraw && '🤝 IT\'S A DRAW! 🤝'}
            {!isWinner && !isDraw && '😞 YOU LOST 😞'}
          </Text>
        </View>

        {/* Stats */}
        <View style={styles.statsContainer}>
          {/* Your Stats */}
          <View style={styles.statsPanel}>
            <Text style={styles.statsTitle}>Your Stats</Text>
            <View style={styles.statRow}>
              <Text style={styles.statLabel}>Rounds Won:</Text>
              <Text style={styles.statValue}>{gameState.yourScore}</Text>
            </View>
            <View style={styles.statRow}>
              <Text style={styles.statLabel}>Rounds Lost:</Text>
              <Text style={styles.statValue}>{gameState.maxRounds - gameState.yourScore}</Text>
            </View>
            <View style={styles.statRow}>
              <Text style={styles.statLabel}>Total Latency:</Text>
              <Text style={styles.statValue}>{gameState.yourLatency}ms</Text>
            </View>
            {gameState.yourScore > 0 && (
              <View style={styles.statRow}>
                <Text style={styles.statLabel}>Avg Latency:</Text>
                <Text style={styles.statValue}>
                  {Math.round(gameState.yourLatency / gameState.yourScore)}ms
                </Text>
              </View>
            )}
          </View>

          {/* Opponent Stats */}
          <View style={styles.statsPanel}>
            <Text style={styles.statsTitle}>Opponent Stats</Text>
            <View style={styles.statRow}>
              <Text style={styles.statLabel}>Rounds Won:</Text>
              <Text style={styles.statValue}>{gameState.opponentScore}</Text>
            </View>
            <View style={styles.statRow}>
              <Text style={styles.statLabel}>Rounds Lost:</Text>
              <Text style={styles.statValue}>{gameState.maxRounds - gameState.opponentScore}</Text>
            </View>
            <View style={styles.statRow}>
              <Text style={styles.statLabel}>Total Latency:</Text>
              <Text style={styles.statValue}>{gameState.opponentLatency}ms</Text>
            </View>
            {gameState.opponentScore > 0 && (
              <View style={styles.statRow}>
                <Text style={styles.statLabel}>Avg Latency:</Text>
                <Text style={styles.statValue}>
                  {Math.round(gameState.opponentLatency / gameState.opponentScore)}ms
                </Text>
              </View>
            )}
          </View>
        </View>

        {/* Round History */}
        <View style={styles.historyContainer}>
          <Text style={styles.historyTitle}>Round History</Text>
          {gameState.roundResults.map((result) => (
            <View key={result.round} style={styles.historyRow}>
              <Text style={styles.historyRound}>Round {result.round}</Text>
              <Text style={[
                styles.historyWinner,
                result.winner === userId && styles.historyWin,
              ]}>
                {result.winner === userId ? '✅ You won' : '❌ Opponent won'}
              </Text>
              <Text style={styles.historyLatency}>{result.latency}ms</Text>
            </View>
          ))}
        </View>

        {/* Play Again Button */}
        <TouchableOpacity style={styles.button} onPress={onPlayAgain}>
          <Text style={styles.buttonText}>🎮 Play Again</Text>
        </TouchableOpacity>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#1a1a2e',
  },
  scrollContent: {
    padding: 24,
  },
  header: {
    alignItems: 'center',
    marginBottom: 24,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#fff',
    letterSpacing: 2,
  },
  banner: {
    padding: 24,
    borderRadius: 16,
    marginBottom: 32,
    alignItems: 'center',
  },
  bannerWin: {
    backgroundColor: '#44ff44',
  },
  bannerDraw: {
    backgroundColor: '#ffaa44',
  },
  bannerLose: {
    backgroundColor: '#ff4444',
  },
  bannerText: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#000',
  },
  statsContainer: {
    marginBottom: 32,
  },
  statsPanel: {
    backgroundColor: '#16213e',
    borderRadius: 12,
    padding: 20,
    marginBottom: 16,
  },
  statsTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#e94560',
    marginBottom: 16,
    textAlign: 'center',
  },
  statRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#0f3460',
  },
  statLabel: {
    fontSize: 16,
    color: '#aaa',
  },
  statValue: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#fff',
  },
  historyContainer: {
    backgroundColor: '#16213e',
    borderRadius: 12,
    padding: 20,
    marginBottom: 32,
  },
  historyTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#e94560',
    marginBottom: 16,
    textAlign: 'center',
  },
  historyRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#0f3460',
  },
  historyRound: {
    fontSize: 14,
    color: '#aaa',
    flex: 1,
  },
  historyWinner: {
    fontSize: 14,
    color: '#ff4444',
    flex: 2,
    textAlign: 'center',
  },
  historyWin: {
    color: '#44ff44',
  },
  historyLatency: {
    fontSize: 14,
    color: '#aaa',
    flex: 1,
    textAlign: 'right',
  },
  button: {
    backgroundColor: '#e94560',
    borderRadius: 12,
    padding: 18,
    alignItems: 'center',
  },
  buttonText: {
    color: '#fff',
    fontSize: 20,
    fontWeight: 'bold',
  },
});