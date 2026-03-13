<?php

namespace App\Logging;

use Psr\Log\LoggerInterface;

/**
 * Collects key=value pairs throughout a request and emits them as a single
 * canonical log line — a compact, grep-friendly format popularised by Stripe.
 */
class CanonicalLogger
{
    /** @var array<string, mixed> */
    private array $fields = [];

    private ?LoggerInterface $logger;

    public function __construct(?LoggerInterface $logger = null)
    {
        $this->logger = $logger;
    }

    public function add(string $key, mixed $value): void
    {
        $this->fields[$key] = $value;
    }

    /** @param array<string, mixed> $pairs */
    public function addMany(array $pairs): void
    {
        foreach ($pairs as $key => $value) {
            $this->fields[$key] = $value;
        }
    }

    public function emit(): void
    {
        $parts = [];
        foreach ($this->fields as $key => $value) {
            $parts[] = $key . '=' . $this->formatValue($value);
        }

        $line = implode(' ', $parts);

        if ($this->logger) {
            $this->logger->info($line);
        } else {
            error_log($line);
        }

        $this->fields = [];
    }

    private function formatValue(mixed $value): string
    {
        if ($value === null) {
            return 'null';
        }

        $str = (string) $value;

        if (str_contains($str, ' ') || $str === '') {
            return '"' . addcslashes($str, '"\\') . '"';
        }

        return $str;
    }
}
