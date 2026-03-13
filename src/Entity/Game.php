<?php

namespace App\Entity;

class Game
{
    public string $id;
    public ?string $player1Move = null;
    public ?string $player2Move = null;
    /** 'waiting' | 'complete' */
    public string $status = 'waiting';
    /** 'player1' | 'player2' | 'draw' | null */
    public ?string $winner = null;
    public string $createdAt;

    public function toArray(): array
    {
        return [
            'id'           => $this->id,
            'player1_move' => $this->player1Move,
            'player2_move' => $this->player2Move,
            'status'       => $this->status,
            'winner'       => $this->winner,
            'created_at'   => $this->createdAt,
        ];
    }
}
