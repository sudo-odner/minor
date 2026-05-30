CREATE OR REPLACE FUNCTION clean_deleted_role_overrides()
RETURNS TRIGGER AS $$
BEGIN
    DELETE FROM channel_permission_overrides WHERE target_type = 'role' AND target_id = OLD.id;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_after_delete_role
AFTER DELETE ON roles
FOR EACH ROW
EXECUTE FUNCTION clean_deleted_role_overrides();
