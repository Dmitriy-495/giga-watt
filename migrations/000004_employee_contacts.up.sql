-- Giga Watt
-- Iteration 001 — Foundation
--
-- "Для каждого типа контакта допускается основной контакт" (ITERATION.md).
-- Это означает не более одного основного телефона и не более одного
-- основного e-mail на сотрудника. Частичный уникальный индекс защищает
-- это на уровне БД, а не только в backend.

CREATE UNIQUE INDEX uq_employee_phones_primary
    ON employee_phones (employee_id)
    WHERE is_primary;

CREATE UNIQUE INDEX uq_employee_emails_primary
    ON employee_emails (employee_id)
    WHERE is_primary;
