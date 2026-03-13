<?php

namespace App\Controller;

use App\Service\GameService;
use CanonicalLogs\CanonicalLogger;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\Routing\Attribute\Route;

class GameController extends AbstractController
{
    public function __construct(private readonly GameService $gameService) {}

    #[Route('/games', name: 'create_game', methods: ['POST'])]
    public function createGame(Request $request): JsonResponse
    {
        $start  = microtime(true);
        $logger = new CanonicalLogger();
        $logger->addMany([
            'action'     => 'create_game',
            'request_id' => $request->headers->get('X-Request-ID', uniqid('req_')),
        ]);

        try {
            $game = $this->gameService->createGame();

            $logger->addMany([
                'game_id'     => $game->id,
                'http_status' => 201,
                'duration_ms' => $this->ms($start),
            ]);

            return $this->json($game->toArray(), 201);
        } catch (\Throwable $e) {
            $logger->addMany([
                'http_status' => 500,
                'error'       => $e->getMessage(),
                'duration_ms' => $this->ms($start),
            ]);

            return $this->json(['error' => 'Failed to create game.'], 500);
        } finally {
            $logger->emit();
        }
    }

    #[Route('/games/{id}/moves', name: 'make_move', methods: ['POST'])]
    public function makeMove(string $id, Request $request): JsonResponse
    {
        $start  = microtime(true);
        $logger = new CanonicalLogger();
        $logger->addMany([
            'action'     => 'make_move',
            'request_id' => $request->headers->get('X-Request-ID', uniqid('req_')),
            'game_id'    => $id,
        ]);

        $body   = json_decode($request->getContent(), associative: true) ?? [];
        $player = $body['player'] ?? null;
        $choice = $body['choice'] ?? null;

        $logger->addMany([
            'player' => $player,
            'choice' => $choice,
        ]);

        try {
            $game = $this->gameService->makeMove($id, (int) $player, (string) $choice);

            $logger->addMany([
                'game_status' => $game->status,
                'winner'      => $game->winner,
                'http_status' => 200,
                'duration_ms' => $this->ms($start),
            ]);

            return $this->json($game->toArray());
        } catch (\InvalidArgumentException $e) {
            $logger->addMany([
                'http_status' => 422,
                'error'       => $e->getMessage(),
                'duration_ms' => $this->ms($start),
            ]);

            return $this->json(['error' => $e->getMessage()], 422);
        } catch (\RuntimeException $e) {
            $status = str_contains($e->getMessage(), 'not found') ? 404 : 500;

            $logger->addMany([
                'http_status' => $status,
                'error'       => $e->getMessage(),
                'duration_ms' => $this->ms($start),
            ]);

            $message = $status === 404 ? 'Game not found.' : 'Unexpected error.';
            return $this->json(['error' => $message], $status);
        } finally {
            $logger->emit();
        }
    }

    private function ms(float $start): float
    {
        return round((microtime(true) - $start) * 1000, 2);
    }
}
