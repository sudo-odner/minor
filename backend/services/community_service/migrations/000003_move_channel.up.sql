CREATE FUNCTION move_channel(
    p_server_id     UUID,
    p_channel_id    UUID,
    p_old_parent_id UUID,
    p_new_parent_id UUID,
    p_old_pos       INTEGER,
    p_new_pos       INTEGER
) RETURNS VOID AS $$
DECLARE
    v_is_same_parent BOOLEAN;
BEGIN
    v_is_same_parent := (p_old_parent_id IS NOT DISTINCT FROM p_new_parent_id);

    if v_is_same_parent THEN
        -- Перемещение в нутри родитлея(категории)
        IF p_old_pos > p_new_pos THEN
            -- Сдвигаем всех те кто ниже вниз
            UPDATE channels 
            SET position = position + 1
            WHERE server_id = p_server_id
                AND parent_id IS NOT DISTINCT FROM p_old_parent_id
                AND position >= p_new_pos
                AND position < p_old_pos;
        ELSIF p_old_pos < p_new_pos THEN
            -- Сдвигаем всех те кто выше вверх
            UPDATE channels
            SET position = position - 1
            WHERE server_id = p_server_id
                AND parent_id IS NOT DISTINCT FROM p_old_parent_id
                AND position > p_old_pos
                AND position <= p_new_pos;
        END IF;
    ELSE 
        -- Перенос в другую категорию (смена родителя)

        -- Обновляем позиции в предыдущей категории(родителя)
        UPDATE channels
        SET position = position - 1
        WHERE server_id = p_server_id
            AND parent_id IS NOT DISTINCT FROM p_old_parent_id
            AND position > p_old_pos;
        
        -- Добавляем место для канала у нового родитея (сдвигаем элементы)
        UPDATE channels
        SET position = position + 1
        WHERE server_id = p_server_id
            AND parent_id IS NOT DISTINCT FROM p_new_parent_id
            AND position >= p_new_pos;
    END IF;

    -- Обновляем у канала position на новые
    UPDATE channels
    SET parent_id = p_new_parent_id, position = p_new_pos
    WHERE id = p_channel_id AND server_id = p_server_id;
EnD;
$$ LANGUAGE plpgsql;


