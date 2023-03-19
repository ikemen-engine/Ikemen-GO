-- This file is for making extra functions that aren't
-- in Lua by default and would be very useful.

-- * table.getKey(tbl, value): Gets the first key that corresponds to
--   the value passed into the function. Helpful for removing something
--   from a table if the key may have changed, or if the key is unknown.
function table.getKey(tbl, value)
    for i,k in pairs(tbl) do
        if k == value then
            return i
        end
    end
end