<?php

namespace App\Service;

use App\Entity\Game;
use Symfony\Component\HttpKernel\KernelInterface;

class GameService
{
    private string $storageFile;

    private const VALID_MOVES = ['rock', 'paper', 'scissors'];

    // Indexed by [attacker][defender] → true means attacker wins
    private const BEATS = [
        'rock'     => ['scissors' => true],
        'scissors' => ['paper'    => true],
        'paper'    => ['rock'     => true],
    ];

    public function __construct(KernelInterface $kernel)
    {
        $dataDir = $kernel->getProjectDir() . '/var/data';
        if (!is_dir($dataDir)) {
            mkdir($dataDir, 0755, recursive: true);
        }
        $this->storageFile = $dataDir . '/games.json';
    }

    public function createGame(): Game
    {
        $game            = new Game();
        $game->id        = bin2hex(random_bytes(8));
        $game->createdAt = (new \DateTimeImmutable())->format(\DateTimeInterface::ATOM);

        $games           = $this->loadGames();
        $games[$game->id] = $game->toArray();
        $this->saveGames($games);

        return $game;
    }

    /** @throws \InvalidArgumentException on bad input */
    public function makeMove(string $gameId, int $player, string $choice): Game
    {
        if (!in_array($player, [1, 2], strict: true)) {
            throw new \InvalidArgumentException('Player must be 1 or 2.');
        }

        if (!in_array($choice, self::VALID_MOVES, strict: true)) {
            throw new \InvalidArgumentException(
                'Choice must be one of: ' . implode(', ', self::VALID_MOVES) . '.'
            );
        }

        $games = $this->loadGames();

        if (!isset($games[$gameId])) {
            throw new \RuntimeException('Game not found.');
        }

        $data = $games[$gameId];

        if ($data['status'] === 'complete') {
            throw new \InvalidArgumentException('Game is already complete.');
        }

        $moveField = $player === 1 ? 'player1_move' : 'player2_move';

        if ($data[$moveField] !== null) {
            throw new \InvalidArgumentException("Player {$player} has already moved.");
        }

        $data[$moveField] = $choice;

        if ($data['player1_move'] !== null && $data['player2_move'] !== null) {
            $data['status'] = 'complete';
            $data['winner'] = $this->determineWinner($data['player1_move'], $data['player2_move']);
        }

        $games[$gameId] = $data;
        $this->saveGames($games);

        return $this->hydrate($data);
    }

    private function determineWinner(string $p1, string $p2): string
    {
        if ($p1 === $p2) {
            return 'draw';
        }
        return (self::BEATS[$p1][$p2] ?? false) ? 'player1' : 'player2';
    }

    private function hydrate(array $data): Game
    {
        $game              = new Game();
        $game->id          = $data['id'];
        $game->player1Move = $data['player1_move'];
        $game->player2Move = $data['player2_move'];
        $game->status      = $data['status'];
        $game->winner      = $data['winner'];
        $game->createdAt   = $data['created_at'];
        return $game;
    }

    private function loadGames(): array
    {
        if (!file_exists($this->storageFile)) {
            return [];
        }
        return json_decode(file_get_contents($this->storageFile), associative: true) ?? [];
    }

    private function saveGames(array $games): void
    {
        file_put_contents($this->storageFile, json_encode($games, JSON_PRETTY_PRINT));
    }
}
